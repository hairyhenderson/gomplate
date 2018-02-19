.DEFAULT_GOAL = build
extension = $(patsubst windows,.exe,$(filter windows,$(1)))
GO := go
PKG_NAME := gomplate
PREFIX := .

COMMIT ?= `git rev-parse --short HEAD 2>/dev/null`
VERSION ?= `git describe --abbrev=0 --tags $(git rev-list --tags --max-count=1) 2>/dev/null | sed 's/v\(.*\)/\1/'`
BUILD_DATE ?= `date -u +"%Y-%m-%dT%H:%M:%SZ"`

COMMIT_FLAG := -X `go list ./version`.GitCommit=$(COMMIT)
VERSION_FLAG := -X `go list ./version`.Version=$(VERSION)

GOOS ?= $(shell go version | sed 's/^.*\ \([a-z0-9]*\)\/\([a-z0-9]*\)/\1/')
GOARCH ?= $(shell go version | sed 's/^.*\ \([a-z0-9]*\)\/\([a-z0-9]*\)/\2/')

platforms := linux-amd64 linux-386 linux-arm linux-arm64 darwin-amd64 solaris-amd64 windows-amd64.exe windows-386.exe
compressed-platforms := linux-amd64-slim linux-arm-slim linux-arm64-slim darwin-amd64-slim windows-amd64-slim.exe

clean:
	rm -Rf $(PREFIX)/bin/*
	rm -f $(PREFIX)/*.[ci]id
	rm -f $(PREFIX)/test/integration/gomplate
	rm -f $(PREFIX)/test/integration/mirror
	rm -f $(PREFIX)/test/integration/meta
	rm -f $(PREFIX)/test/integration/aws

build-x: $(patsubst %,$(PREFIX)/bin/$(PKG_NAME)_%,$(platforms))

compress-all: $(patsubst %,$(PREFIX)/bin/$(PKG_NAME)_%,$(compressed-platforms))

$(PREFIX)/bin/$(PKG_NAME)_%-slim: $(PREFIX)/bin/$(PKG_NAME)_%
	upx --lzma $< -o $@

$(PREFIX)/bin/$(PKG_NAME)_%-slim.exe: $(PREFIX)/bin/$(PKG_NAME)_%.exe
	upx --lzma $< -o $@

compress: $(PREFIX)/bin/$(PKG_NAME)_$(GOOS)-$(GOARCH)-slim$(call extension,$(GOOS))
	cp $< $(PREFIX)/bin/$(PKG_NAME)-slim$(call extension,$(GOOS))

%.iid: Dockerfile
	@docker build \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		--build-arg VCS_REF=$(COMMIT) \
		--target $(subst .iid,,$@) \
		--iidfile $@ \
		.

%.cid: %.iid
	@docker create $(shell cat $<) > $@

build-release: artifacts.cid
	@docker cp $(shell cat $<):/bin/. bin/

docker-images: gomplate.iid gomplate-slim.iid

$(PREFIX)/bin/$(PKG_NAME)_%: $(shell find $(PREFIX) -type f -name '*.go' -not -path "$(PREFIX)/test/*")
	GOOS=$(shell echo $* | cut -f1 -d-) GOARCH=$(shell echo $* | cut -f2 -d- | cut -f1 -d.) CGO_ENABLED=0 \
		$(GO) build \
			-ldflags "-w -s $(COMMIT_FLAG) $(VERSION_FLAG)" \
			-o $@

$(PREFIX)/bin/$(PKG_NAME)$(call extension,$(GOOS)): $(PREFIX)/bin/$(PKG_NAME)_$(GOOS)-$(GOARCH)$(call extension,$(GOOS))
	cp $< $@

$(PREFIX)/test/integration/mirror$(call extension,$(GOOS)): $(shell find $(PREFIX)/test/integration/mirrorsvc -type f -name '*.go')
	CGO_ENABLED=0 \
		$(GO) build -ldflags "-w -s $(COMMIT_FLAG) $(VERSION_FLAG)" -o $@ $(PREFIX)/test/integration/mirrorsvc

$(PREFIX)/test/integration/meta$(call extension,$(GOOS)): $(shell find $(PREFIX)/test/integration/metasvc -type f -name '*.go')
	CGO_ENABLED=0 \
		$(GO) build -ldflags "-w -s $(COMMIT_FLAG) $(VERSION_FLAG)" -o $@ $(PREFIX)/test/integration/metasvc

$(PREFIX)/test/integration/aws$(call extension,$(GOOS)): $(shell find $(PREFIX)/test/integration/awssvc -type f -name '*.go')
	CGO_ENABLED=0 \
		$(GO) build -ldflags "-w -s $(COMMIT_FLAG) $(VERSION_FLAG)" -o $@ $(PREFIX)/test/integration/awssvc

build: $(PREFIX)/bin/$(PKG_NAME)$(call extension,$(GOOS))

build-mirror: $(PREFIX)/test/integration/mirror$(call extension,$(GOOS))

build-meta: $(PREFIX)/test/integration/meta$(call extension,$(GOOS))

build-aws: $(PREFIX)/test/integration/aws$(call extension,$(GOOS))

test:
	$(GO) test -v -race ./...

build-integration-image: $(PREFIX)/bin/$(PKG_NAME)_linux-amd64$(call extension,$(GOOS)) $(PREFIX)/bin/mirror_linux-amd64$(call extension,$(GOOS)) $(PREFIX)/bin/meta_linux-amd64$(call extension,$(GOOS)) $(PREFIX)/bin/aws_linux-amd64$(call extension,$(GOOS))
	cp $(PREFIX)/bin/$(PKG_NAME)_linux-amd64 test/integration/gomplate
	cp $(PREFIX)/bin/mirror_linux-amd64 test/integration/mirror
	cp $(PREFIX)/bin/meta_linux-amd64 test/integration/meta
	cp $(PREFIX)/bin/aws_linux-amd64 test/integration/aws
	docker build -f test/integration/Dockerfile -t gomplate-test test/integration/

test-integration-docker: build-integration-image
	docker run -it --rm gomplate-test

test-integration: build build-mirror build-meta build-aws
	@test/integration/test.sh

gen-changelog:
	docker run -it -v $(pwd):/app --workdir /app -e CHANGELOG_GITHUB_TOKEN hairyhenderson/github_changelog_generator \
		github_changelog_generator --no-filter-by-milestone --exclude-labels duplicate,question,invalid,wontfix,admin

docs/themes/hugo-material-docs:
	git clone https://github.com/digitalcraftsman/hugo-material-docs.git $@

gen-docs: docs/themes/hugo-material-docs
	cd docs/; hugo

# this target doesn't usually get used - it's mostly here as a reminder to myself
# hint: make sure CLOUDCONVERT_API_KEY is set ;)
gomplate.png: gomplate.svg
	cloudconvert -f png -c density=288 $^

ifeq ("$(CI)","true")
lint:
	gometalinter -j 1 --vendor --deadline 120s --disable gotype --enable gofmt --enable goimports --enable misspell --enable unused --disable gas
else
lint:
	gometalinter -j $(shell nproc) --vendor --deadline 120s --disable gotype --enable gofmt --enable goimports --enable misspell --enable unused --disable gas
endif

.PHONY: gen-changelog clean test build-x compress-all build-release build build-integration-image test-integration-docker gen-docs lint clean-images clean-containers docker-images
.DELETE_ON_ERROR:
.SECONDARY:
