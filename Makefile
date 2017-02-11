extension = $(patsubst windows,.exe,$(filter windows,$(1)))
GO := go
PKG_NAME := gomplate
PREFIX := .

COMMIT ?= `git rev-parse --short HEAD 2>/dev/null`
VERSION ?= `git describe --abbrev=0 --tags $(git rev-list --tags --max-count=1) 2>/dev/null | sed 's/v\(.*\)/\1/'`

COMMIT_FLAG := -X `go list ./version`.GitCommit=$(COMMIT)
VERSION_FLAG := -X `go list ./version`.Version=$(VERSION)

define gocross
	GOOS=$(1) GOARCH=$(2) \
		$(GO) build \
			-ldflags "$(COMMIT_FLAG) $(VERSION_FLAG)" \
			-o $(PREFIX)/bin/$(PKG_NAME)_$(1)-$(2)$(call extension,$(1));
endef

define compress
	upx $(PREFIX)/bin/$(PKG_NAME)_$(1)-$(2)$(call extension,$(1)) \
	 -o $(PREFIX)/bin/$(PKG_NAME)_$(1)-$(2)-slim$(call extension,$(1))
endef

clean:
	rm -Rf $(PREFIX)/bin/*

build-x: $(shell find . -type f -name '*.go')
	$(call gocross,linux,amd64)
	$(call gocross,linux,386)
	$(call gocross,linux,arm)
	$(call gocross,darwin,amd64)
	$(call gocross,windows,amd64)
	$(call gocross,windows,386)

compress-all:
	$(call compress,linux,amd64)
	$(call compress,linux,386)
	$(call compress,linux,arm)
	$(call compress,darwin,amd64)
	$(call compress,windows,amd64)
	$(call compress,windows,386)

build-release: clean build-x compress-all

$(PREFIX)/bin/$(PKG_NAME)$(call extension,$(GOOS)): $(shell find . -type f -name '*.go')
	$(GO) build -ldflags "$(COMMIT_FLAG) $(VERSION_FLAG)" -o $@

build: $(PREFIX)/bin/$(PKG_NAME)$(call extension,$(GOOS))

test:
	$(GO) test -v -race `glide novendor`

gen-changelog:
	github_changelog_generator --no-filter-by-milestone --exclude-labels duplicate,question,invalid,wontfix,admin

.PHONY: gen-changelog
