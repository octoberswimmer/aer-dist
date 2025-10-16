VERSION := $(shell git describe --tags --abbrev=0 --always --dirty 2>/dev/null || echo dev)
EXECUTABLE := aer
WINDOWS := $(EXECUTABLE)_windows_amd64.exe
LINUX := $(EXECUTABLE)_linux_amd64
LINUX_ARM64 := $(EXECUTABLE)_linux_arm64
DARWIN_AMD64 := $(EXECUTABLE)_darwin_amd64
DARWIN_ARM64 := $(EXECUTABLE)_darwin_arm64
ALL := $(WINDOWS) $(LINUX) $(LINUX_ARM64) $(DARWIN_AMD64) $(DARWIN_ARM64)

GO_BUILD_FLAGS := -trimpath

.PHONY: default install install-debug dist clean checksum

default:
	go build $(GO_BUILD_FLAGS)

install:
	go install $(GO_BUILD_FLAGS)

install-debug:
	go install $(GO_BUILD_FLAGS) -gcflags="all=-N -l"

$(WINDOWS):
	env CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(GO_BUILD_FLAGS) -o $@

$(LINUX):
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(GO_BUILD_FLAGS) -o $@

$(LINUX_ARM64):
	env CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build $(GO_BUILD_FLAGS) -o $@

$(DARWIN_AMD64):
	env CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build $(GO_BUILD_FLAGS) -o $@

$(DARWIN_ARM64):
	env CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build $(GO_BUILD_FLAGS) -o $@

$(basename $(WINDOWS)).zip: $(WINDOWS)
	zip $@ $<
	7za rn $@ $< $(EXECUTABLE).exe

%.zip: %
	zip $@ $<
	7za rn $@ $< $(EXECUTABLE)

dist: $(addsuffix .zip,$(basename $(ALL)))

checksum: dist
	shasum -a 256 $(addsuffix .zip,$(basename $(ALL))) > SHA256SUMS-$(VERSION)

clean:
	-rm -f $(EXECUTABLE) $(EXECUTABLE).exe $(EXECUTABLE)_* *.zip SHA256SUMS-*

