.DEFAULT_GOAL = build
extension = $(patsubst windows,.exe,$(filter windows,$(1)))
GO := go
PKG_NAME := gomplate
PREFIX := .
GOFLAGS := -mod=vendor

ifeq ("$(CI)","true")
LINT_PROCS ?= 1
else
LINT_PROCS ?= $(shell nproc)
endif

COMMIT ?= `git rev-parse --short HEAD 2>/dev/null`
VERSION ?= `git describe --abbrev=0 --tags $(git rev-list --tags --max-count=1) 2>/dev/null | sed 's/v\(.*\)/\1/'`

COMMIT_FLAG := -X `go list ./version`.GitCommit=$(COMMIT)
VERSION_FLAG := -X `go list ./version`.Version=$(VERSION)

GOOS ?= $(shell go version | sed 's/^.*\ \([a-z0-9]*\)\/\([a-z0-9]*\)/\1/')
GOARCH ?= $(shell go version | sed 's/^.*\ \([a-z0-9]*\)\/\([a-z0-9]*\)/\2/')

platforms := freebsd-amd64 linux-amd64 linux-386 linux-arm linux-arm64 darwin-amd64 solaris-amd64 windows-amd64.exe windows-386.exe
compressed-platforms := linux-amd64-slim linux-arm-slim linux-arm64-slim darwin-amd64-slim windows-amd64-slim.exe

clean:
	rm -Rf $(PREFIX)/bin/*
	rm -f $(PREFIX)/*.[ci]id

build-x: $(patsubst %,$(PREFIX)/bin/$(PKG_NAME)_%,$(platforms))

compress-all: $(patsubst %,$(PREFIX)/bin/$(PKG_NAME)_%,$(compressed-platforms))

$(PREFIX)/bin/$(PKG_NAME)_%-slim: $(PREFIX)/bin/$(PKG_NAME)_%
	upx --lzma $< -o $@

$(PREFIX)/bin/$(PKG_NAME)_%-slim.exe: $(PREFIX)/bin/$(PKG_NAME)_%.exe
	upx --lzma $< -o $@

$(PREFIX)/bin/$(PKG_NAME)_%_checksum.txt: $(PREFIX)/bin/$(PKG_NAME)_%
	@sha256sum $< > $@

$(PREFIX)/bin/checksums.txt: \
		$(patsubst %,$(PREFIX)/bin/$(PKG_NAME)_%_checksum.txt,$(platforms)) \
		$(patsubst %,$(PREFIX)/bin/$(PKG_NAME)_%_checksum.txt,$(compressed-platforms))
	@cat $^ > $@

$(PREFIX)/%.signed: $(PREFIX)/%
	@keybase sign < $< > $@

compress: $(PREFIX)/bin/$(PKG_NAME)_$(GOOS)-$(GOARCH)-slim$(call extension,$(GOOS))
	cp $< $(PREFIX)/bin/$(PKG_NAME)-slim$(call extension,$(GOOS))

%.iid: Dockerfile
	@docker build \
		--build-arg VCS_REF=$(COMMIT) \
		--target $(subst .iid,,$@) \
		--iidfile $@ \
		.

%.cid: %.iid
	@docker create $(shell cat $<) > $@

build-release: artifacts.cid
	@docker cp $(shell cat $<):/bin/. bin/

docker-images: gomplate.iid gomplate-slim.iid

$(PREFIX)/bin/$(PKG_NAME)_%: $(shell find $(PREFIX) -type f -name "*.go")
	GOOS=$(shell echo $* | cut -f1 -d-) GOARCH=$(shell echo $* | cut -f2 -d- | cut -f1 -d.) CGO_ENABLED=0 \
		GOFLAGS=$(GOFLAGS) \
		$(GO) build \
			-ldflags "-w -s $(COMMIT_FLAG) $(VERSION_FLAG)" \
			-o $@ \
			./cmd/gomplate

$(PREFIX)/bin/$(PKG_NAME)$(call extension,$(GOOS)): $(PREFIX)/bin/$(PKG_NAME)_$(GOOS)-$(GOARCH)$(call extension,$(GOOS))
	cp $< $@

build: $(PREFIX)/bin/$(PKG_NAME)$(call extension,$(GOOS))

ifeq ($(OS),Windows_NT)
test:
	@GOFLAGS=$(GOFLAGS) $(GO) test -v -coverprofile=c.out ./...
else
test:
	@GOFLAGS=$(GOFLAGS) $(GO) test -v -race -coverprofile=c.out ./...
endif

integration: build
	@GOFLAGS=$(GOFLAGS) $(GO) test -v -tags=integration \
		./tests/integration

integration.iid: Dockerfile.integration $(PREFIX)/bin/$(PKG_NAME)_linux-amd64$(call extension,$(GOOS))
	docker build -f $< --iidfile $@ .

test-integration-docker: integration.iid
	docker run -it --rm $(shell cat $<)

gen-changelog:
	docker run -it -v $(shell pwd):/app --workdir /app -e CHANGELOG_GITHUB_TOKEN hairyhenderson/github_changelog_generator \
		github_changelog_generator --no-filter-by-milestone --exclude-labels duplicate,question,invalid,wontfix,admin

docs/themes/hugo-material-docs:
	git clone https://github.com/digitalcraftsman/hugo-material-docs.git $@

gen-docs: docs/themes/hugo-material-docs
	cd docs/; hugo

docs/content/functions/%.md: docs-src/content/functions/%.yml docs-src/content/functions/func_doc.md.tmpl
	gomplate -d data=$< -f docs-src/content/functions/func_doc.md.tmpl -o $@

# this target doesn't usually get used - it's mostly here as a reminder to myself
# hint: make sure CLOUDCONVERT_API_KEY is set ;)
gomplate.png: gomplate.svg
	cloudconvert -f png -c density=288 $^

lint:
	@golangci-lint run --timeout 2m --disable-all \
		--enable depguard \
		--enable dupl \
		--enable goconst \
		--enable gocritic \
		--enable gocyclo \
		--enable gofmt \
		--enable goimports \
		--enable golint \
		--enable gosec \
		--enable gosimple \
		--enable govet \
		--enable ineffassign \
		--enable maligned \
		--enable misspell \
		--enable nakedret \
		--enable prealloc \
		--enable staticcheck \
		--enable structcheck \
		--enable stylecheck \
		--enable typecheck \
		--enable unconvert \
		--enable varcheck

	@golangci-lint run --timeout 2m --tests=false --disable-all \
		--enable deadcode \
		--enable errcheck \
		--enable interfacer \
		--enable scopelint \
		--enable unused

	@golangci-lint run --timeout 2m --build-tags integration \
		--disable-all \
		--enable deadcode \
		--enable depguard \
		--enable dupl \
		--enable gochecknoinits \
		--enable gocritic \
		--enable gocyclo \
		--enable gofmt \
		--enable goimports \
		--enable golint \
		--enable gosec \
		--enable gosimple \
		--enable govet \
		--enable ineffassign \
		--enable maligned \
		--enable misspell \
		--enable nakedret \
		--enable prealloc \
		--enable scopelint \
		--enable staticcheck \
		--enable structcheck \
		--enable stylecheck \
		--enable typecheck \
		--enable unconvert \
		--enable unparam \
		--enable unused \
		--enable varcheck \
			./tests/integration

.PHONY: gen-changelog clean test build-x compress-all build-release build test-integration-docker gen-docs lint clean-images clean-containers docker-images
.DELETE_ON_ERROR:
.SECONDARY:
