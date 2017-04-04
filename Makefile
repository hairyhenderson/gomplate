extension = $(patsubst windows,.exe,$(filter windows,$(1)))
GO := go
PKG_NAME := gomplate
PREFIX := .

COMMIT ?= `git rev-parse --short HEAD 2>/dev/null`
VERSION ?= `git describe --abbrev=0 --tags $(git rev-list --tags --max-count=1) 2>/dev/null | sed 's/v\(.*\)/\1/'`

COMMIT_FLAG := -X `go list ./version`.GitCommit=$(COMMIT)
VERSION_FLAG := -X `go list ./version`.Version=$(VERSION)

platforms := linux-amd64 linux-386 linux-arm linux-arm64 darwin-amd64 windows-amd64.exe windows-386.exe

define gocross
	GOOS=$(1) GOARCH=$(2) \
		$(GO) build \
			-ldflags "-w -s $(COMMIT_FLAG) $(VERSION_FLAG)" \
			-o $(PREFIX)/bin/$(PKG_NAME)_$(1)-$(2)$(call extension,$(1));
endef

define compress
	upx $(PREFIX)/bin/$(PKG_NAME)_$(1)-$(2)$(call extension,$(1)) \
	 -o $(PREFIX)/bin/$(PKG_NAME)_$(1)-$(2)-slim$(call extension,$(1))
endef

clean:
	rm -Rf $(PREFIX)/bin/*
	rm -f $(PREFIX)/test/integration/gomplate

build-x: $(patsubst %,$(PREFIX)/bin/$(PKG_NAME)_%,$(platforms))

compress-all:
	$(call compress,linux,amd64)
	$(call compress,linux,arm)
	$(call compress,windows,amd64)

build-release: clean build-x compress-all

$(PREFIX)/bin/$(PKG_NAME)_%: $(shell find . -type f -name '*.go')
	$(call gocross,$(shell echo $* | sed 's/\([^-]*\)-\([^.]*\).*/\1/'),$(shell echo $* | sed 's/\([^-]*\)-\([^.]*\).*/\2/'))

$(PREFIX)/bin/$(PKG_NAME)$(call extension,$(GOOS)): $(shell find . -type f -name '*.go')
	$(GO) build -ldflags "-w -s $(COMMIT_FLAG) $(VERSION_FLAG)" -o $@

build: $(PREFIX)/bin/$(PKG_NAME)$(call extension,$(GOOS))

test:
	$(GO) test -v -race `glide novendor`

build-integration-image: $(PREFIX)/bin/$(PKG_NAME)_linux-amd64$(call extension,$(GOOS))
	cp $(PREFIX)/bin/$(PKG_NAME)_linux-amd64 test/integration/gomplate
	docker build -f test/integration/Dockerfile -t gomplate-test test/integration/

test-integration-docker: build-integration-image
	docker run -it --rm gomplate-test

test-integration: build
	@test/integration/test.sh

gen-changelog:
	github_changelog_generator --no-filter-by-milestone --exclude-labels duplicate,question,invalid,wontfix,admin

.PHONY: gen-changelog clean test build-x compress-all build-release build build-integration-image test-integration-docker
