extension = $(patsubst windows,.exe,$(filter windows,$(1)))
GO := go
PKG_NAME := gomplate
PREFIX := .

define gocross
	GOOS=$(1) GOARCH=$(2) \
		$(GO) build \
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
	$(GO) build -o $@

build: $(PREFIX)/bin/$(PKG_NAME)$(call extension,$(GOOS))

gen-changelog:
	github_changelog_generator --exclude-labels duplicate,question,invalid,wontfix,admin

.PHONY: gen-changelog
