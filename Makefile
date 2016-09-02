extension = $(patsubst windows,.exe,$(filter windows,$(1)))
GO := go
PKG_NAME := gomplate
PREFIX := .

COMMIT_FLAG := -X `go list ./version`.GitCommit=`git rev-parse --short HEAD 2>/dev/null`
VERSION_FLAG := -X `go list ./version`.Version=`git describe --abbrev=0 --tags $(git rev-list --tags --max-count=1) 2>/dev/null | sed 's/v\(.*\)/\1/'`

define gocross
	GOOS=$(1) GOARCH=$(2) \
		$(GO) build \
			-ldflags "$(COMMIT_FLAG) $(VERSION_FLAG)" \
			-o $(PREFIX)/bin/$(PKG_NAME)_$(1)-$(2)$(call extension,$(1));
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

$(PREFIX)/bin/$(PKG_NAME)$(call extension,$(GOOS)): $(shell find . -type f -name '*.go')
	$(GO) build -ldflags "$(COMMIT_FLAG) $(VERSION_FLAG)" -o $@

build: $(PREFIX)/bin/$(PKG_NAME)$(call extension,$(GOOS))

gen-changelog:
	github_changelog_generator --no-filter-by-milestone --exclude-labels duplicate,question,invalid,wontfix,admin

.PHONY: gen-changelog
