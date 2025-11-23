SHELL := /bin/bash

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
EXECUTABLE := aer
WINDOWS := $(EXECUTABLE)_windows_amd64.exe
LINUX := $(EXECUTABLE)_linux_amd64
LINUX_ARM64 := $(EXECUTABLE)_linux_arm64
DARWIN_AMD64 := $(EXECUTABLE)_darwin_amd64
DARWIN_ARM64 := $(EXECUTABLE)_darwin_arm64
ALL := $(WINDOWS) $(LINUX) $(LINUX_ARM64) $(DARWIN_AMD64) $(DARWIN_ARM64)
VERSIONED_ZIPS := $(addsuffix _$(VERSION).zip,$(basename $(ALL)))
RELEASE_ASSETS := $(VERSIONED_ZIPS) SHA256SUMS-$(VERSION)

GO_BUILD_FLAGS := -trimpath
GO_LDFLAGS := -X main.version=$(VERSION)

.PHONY: default install install-debug dist clean checksum release tag

default:
	go build $(GO_BUILD_FLAGS) -ldflags "$(GO_LDFLAGS)"

install:
	go install $(GO_BUILD_FLAGS) -ldflags "$(GO_LDFLAGS)"

install-debug:
	go install $(GO_BUILD_FLAGS) -ldflags "$(GO_LDFLAGS)" -gcflags="all=-N -l"

$(WINDOWS): go.mod
	env CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(GO_BUILD_FLAGS) -ldflags "$(GO_LDFLAGS)" -o $@

$(LINUX): go.mod
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(GO_BUILD_FLAGS) -ldflags "$(GO_LDFLAGS)" -o $@

$(LINUX_ARM64): go.mod
	env CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build $(GO_BUILD_FLAGS) -ldflags "$(GO_LDFLAGS)" -o $@

$(DARWIN_AMD64): go.mod
	env CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build $(GO_BUILD_FLAGS) -ldflags "$(GO_LDFLAGS)" -o $@

$(DARWIN_ARM64): go.mod
	env CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build $(GO_BUILD_FLAGS) -ldflags "$(GO_LDFLAGS)" -o $@

$(basename $(WINDOWS))_$(VERSION).zip: $(WINDOWS)
	@rm -f $@
	zip $@ $<
	7za rn $@ $< $(EXECUTABLE).exe

%_$(VERSION).zip: %
	@rm -f $@
	zip $@ $<
	7za rn $@ $< $(EXECUTABLE)

dist: $(VERSIONED_ZIPS)

checksum: dist
	shasum -a 256 $(VERSIONED_ZIPS) > SHA256SUMS-$(VERSION)

release: checksum
	@if ! command -v gh >/dev/null 2>&1; then \
		echo "gh CLI is required for 'make release'."; \
		exit 1; \
	fi
	@if [ "$(VERSION)" = "dev" ] || printf "%s" "$(VERSION)" | grep -q "dirty"; then \
		echo "VERSION '$(VERSION)' is not a clean tag. Tag the commit or invoke 'make release VERSION=vX.Y.Z' with a published tag."; \
		exit 1; \
	fi
	@if ! git rev-parse --verify "refs/tags/$(VERSION)" >/dev/null 2>&1; then \
		echo "Tag '$(VERSION)' does not exist. Create the tag before running 'make release'."; \
		exit 1; \
	fi
	git push octoberswimmer "$(VERSION)"
	gh release create "$(VERSION)" --title "aer $(VERSION)" --notes-from-tag --verify-tag $(RELEASE_ASSETS)

tag:
	@set -euo pipefail; \
	if ! git diff --quiet || ! git diff --cached --quiet; then \
		echo "Working tree must be clean before running 'make tag'."; \
		exit 1; \
	fi; \
	module="github.com/octoberswimmer/aer"; \
	current_version=$$(go list -m -f '{{.Version}}' "$$module"); \
	tags=$$(git ls-remote --tags --refs --sort=v:refname https://github.com/octoberswimmer/aer.git \
		| awk '/refs\/tags\/v/{gsub("refs/tags/","",$$2); print $$2}'); \
	next_version=""; \
	found_state=0; \
	for tag in $$tags; do \
		if [ $$found_state -eq 1 ]; then \
			next_version="$$tag"; \
			break; \
		fi; \
		if [ "$$tag" = "$$current_version" ]; then \
			found_state=1; \
		fi; \
	done; \
	if [ $$found_state -eq 0 ]; then \
		echo "Current version $$current_version not found upstream."; \
		exit 1; \
	fi; \
	if [ -z "$$next_version" ]; then \
		echo "Already on the latest aer tag."; \
		exit 1; \
	fi; \
	echo "Updating $$module from $$current_version to $$next_version"; \
	tmpdir=$$(mktemp -d); \
	trap 'rm -rf "$$tmpdir"' EXIT; \
	git -C "$$tmpdir" init -q; \
	git -C "$$tmpdir" fetch --depth=1 https://github.com/octoberswimmer/aer.git tag "$$next_version" >/dev/null; \
	git -C "$$tmpdir" for-each-ref --format='%(contents)' "refs/tags/$$next_version" > "$$tmpdir/message"; \
	if [ ! -s "$$tmpdir/message" ]; then \
		echo "Release $$next_version" > "$$tmpdir/message"; \
	fi; \
	tag_message_file="$$tmpdir/message"; \
	go get "$$module@$$next_version"; \
	go mod tidy; \
	workflow=".github/workflows/oss-tests.yml"; \
	tmp_workflow=$$(mktemp); \
	if ! sed -e "0,/^\([[:space:]]*AER_VERSION:\) v[0-9][0-9.]*/s//\1 $$next_version/" "$$workflow" > "$$tmp_workflow"; then \
		rm -f "$$tmp_workflow"; \
		exit 1; \
	fi; \
	if cmp -s "$$workflow" "$$tmp_workflow"; then \
		rm -f "$$tmp_workflow"; \
		echo "Failed to update AER_VERSION in $$workflow"; \
		exit 1; \
	fi; \
	mv "$$tmp_workflow" "$$workflow"; \
	$(MAKE) VERSION="$$next_version" dist; \
	git add go.mod go.sum .github/workflows/oss-tests.yml; \
	if git diff --cached --quiet; then \
		echo "No changes produced by the tag workflow."; \
		exit 1; \
	fi; \
	commit_message="Update aer to $$next_version"; \
	git commit -m "$$commit_message"; \
	if git rev-parse --verify --quiet "refs/tags/$$next_version" >/dev/null; then \
		echo "Tag $$next_version already exists locally."; \
		exit 1; \
	fi; \
	git tag -s "$$next_version" -F "$$tag_message_file"; \
	trap - EXIT; \
	rm -rf "$$tmpdir"

test:
	ghproxy --repo octoberswimmer/aer-dist -- act

clean:
	-rm -f $(EXECUTABLE) $(EXECUTABLE).exe $(EXECUTABLE)_* *.zip SHA256SUMS-*
