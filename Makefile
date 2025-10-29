VERSION := $(shell git describe --tags --abbrev=0 --always --dirty 2>/dev/null || echo dev)
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

.PHONY: default install install-debug dist clean checksum release

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
	gh release create "$(VERSION)" --title "aer $(VERSION)" --notes-from-tag --verify-tag $(RELEASE_ASSETS)

clean:
	-rm -f $(EXECUTABLE) $(EXECUTABLE).exe $(EXECUTABLE)_* *.zip SHA256SUMS-*
