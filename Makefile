.DEFAULT_GOAL = build
GO ?= go
extension = $(patsubst windows,.exe,$(filter windows,$(1)))
PKG_NAME := gomplate
DOCKER_REPO ?= hairyhenderson/$(PKG_NAME)
PREFIX := .
DOCKER_LINUX_PLATFORMS ?= linux/amd64,linux/arm64,linux/arm/v6,linux/arm/v7,linux/ppc64le
DOCKER_PLATFORMS ?= $(DOCKER_LINUX_PLATFORMS),windows/amd64
# we just load by default, as a "dry run"
BUILDX_ACTION ?= --load
TAG_LATEST ?= latest
TAG_SLIM ?= slim
TAG_ALPINE ?= alpine

ifeq ("$(CI)","true")
LINT_PROCS ?= 1
else
LINT_PROCS ?= $(shell nproc)
endif

COMMIT ?= `git rev-parse --short HEAD 2>/dev/null`
VERSION ?= `git describe --abbrev=0 --tags $(git rev-list --tags --max-count=1) 2>/dev/null | sed 's/v\(.*\)/\1/'`

COMMIT_FLAG := -X `$(GO) list ./version`.GitCommit=$(COMMIT)
VERSION_FLAG := -X `$(GO) list ./version`.Version=$(VERSION)

GOOS ?= $(shell $(GO) version | sed 's/^.*\ \([a-z0-9]*\)\/\([a-z0-9]*\)/\1/')
GOARCH ?= $(shell $(GO) version | sed 's/^.*\ \([a-z0-9]*\)\/\([a-z0-9]*\)/\2/')

ifeq ("$(TARGETVARIANT)","")
ifneq ("$(GOARM)","")
TARGETVARIANT := v$(GOARM)
endif
else
ifeq ("$(GOARM)","")
GOARM ?= $(subst v,,$(TARGETVARIANT))
endif
endif

# platforms := freebsd-amd64 linux-amd64 linux-386 linux-armv5 linux-armv6 linux-armv7 linux-arm64 darwin-amd64 solaris-amd64 windows-amd64.exe windows-386.exe
platforms := freebsd-amd64 linux-amd64 linux-386 linux-armv6 linux-armv7 linux-arm64 linux-ppc64le darwin-amd64 darwin-arm64 solaris-amd64 windows-amd64.exe windows-386.exe
compressed-platforms := linux-amd64-slim linux-armv6-slim linux-armv7-slim linux-arm64-slim darwin-amd64-slim windows-amd64-slim.exe

clean:
	rm -Rf $(PREFIX)/bin/*
	rm -f $(PREFIX)/*.[ci]id

build-x: $(patsubst %,$(PREFIX)/bin/$(PKG_NAME)_%,$(platforms))

$(PREFIX)/bin/%.zip: $(PREFIX)/bin/%
	@zip -j $@ $^

$(PREFIX)/bin/$(PKG_NAME)_windows-%.zip: $(PREFIX)/bin/$(PKG_NAME)_windows-%.exe
	@zip -j $@ $^

compress-all: $(patsubst %,$(PREFIX)/bin/$(PKG_NAME)_%,$(compressed-platforms))

UPX_VERSION := $(shell upx --version | head -n1 | cut -f2 -d\ )
UPX_REQUIRED_VERSION := 3.94


ifeq ($(UPX_REQUIRED_VERSION),$(UPX_VERSION))
$(PREFIX)/bin/$(PKG_NAME)_%-slim: $(PREFIX)/bin/$(PKG_NAME)_%
	upx --lzma $< -o $@
$(PREFIX)/bin/$(PKG_NAME)_windows-%-slim.exe: $(PREFIX)/bin/$(PKG_NAME)_windows-%.exe
	upx --lzma $< -o $@
else
$(PREFIX)/bin/$(PKG_NAME)_%-slim:
	$(error Wrong upx version - need $(UPX_REQUIRED_VERSION))

$(PREFIX)/bin/$(PKG_NAME)_windows-%-slim.exe:
	$(error Wrong upx version - need $(UPX_REQUIRED_VERSION))
endif


$(PREFIX)/bin/$(PKG_NAME)_%_checksum_sha256.txt: $(PREFIX)/bin/$(PKG_NAME)_%
	@sha256sum $< > $@

$(PREFIX)/bin/$(PKG_NAME)_%_checksum_sha512.txt: $(PREFIX)/bin/$(PKG_NAME)_%
	@sha512sum $< > $@

$(PREFIX)/bin/checksums.txt: $(PREFIX)/bin/checksums_sha256.txt
	@cp $< $@

$(PREFIX)/bin/checksums_sha256.txt: \
		$(patsubst %,$(PREFIX)/bin/$(PKG_NAME)_%_checksum_sha256.txt,$(platforms)) \
		$(patsubst %,$(PREFIX)/bin/$(PKG_NAME)_%_checksum_sha256.txt,$(compressed-platforms))
	@cat $^ > $@

$(PREFIX)/bin/checksums_sha512.txt: \
		$(patsubst %,$(PREFIX)/bin/$(PKG_NAME)_%_checksum_sha512.txt,$(platforms)) \
		$(patsubst %,$(PREFIX)/bin/$(PKG_NAME)_%_checksum_sha512.txt,$(compressed-platforms))
	@cat $^ > $@

$(PREFIX)/%.signed: $(PREFIX)/%
	@keybase sign < $< > $@

compress: $(PREFIX)/bin/$(PKG_NAME)_$(GOOS)-$(GOARCH)$(TARGETVARIANT)-slim$(call extension,$(GOOS))
	cp $< $(PREFIX)/bin/$(PKG_NAME)-slim$(call extension,$(GOOS))

%.iid: Dockerfile
	@docker build \
		--build-arg VCS_REF=$(COMMIT) \
		--target $(subst .iid,,$@) \
		--iidfile $@ \
		.

docker-multi: Dockerfile
	docker buildx build \
		--build-arg VCS_REF=$(COMMIT) \
		--platform $(DOCKER_PLATFORMS) \
		--tag $(DOCKER_REPO):$(TAG_LATEST) \
		--target gomplate \
		$(BUILDX_ACTION) .
	docker buildx build \
		--build-arg VCS_REF=$(COMMIT) \
		--platform $(DOCKER_PLATFORMS) \
		--tag $(DOCKER_REPO):$(TAG_SLIM) \
		--target gomplate-slim \
		$(BUILDX_ACTION) .
	docker buildx build \
		--build-arg VCS_REF=$(COMMIT) \
		--platform $(DOCKER_LINUX_PLATFORMS) \
		--tag $(DOCKER_REPO):$(TAG_ALPINE) \
		--target gomplate-alpine \
		$(BUILDX_ACTION) .

%.cid: %.iid
	@docker create $(shell cat $<) > $@

build-release: artifacts.cid
	@docker cp $(shell cat $<):/bin/. bin/

docker-images: gomplate.iid gomplate-slim.iid

$(PREFIX)/bin/$(PKG_NAME)_%v5$(call extension,$(GOOS)): $(shell find $(PREFIX) -type f -name "*.go") go.mod go.sum
	GOOS=$(shell echo $* | cut -f1 -d-) GOARCH=$(shell echo $* | cut -f2 -d- ) GOARM=5 CGO_ENABLED=0 \
		$(GO) build \
			-ldflags "-w -s $(COMMIT_FLAG) $(VERSION_FLAG)" \
			-o $@ \
			./cmd/$(PKG_NAME)

$(PREFIX)/bin/$(PKG_NAME)_%v6$(call extension,$(GOOS)): $(shell find $(PREFIX) -type f -name "*.go") go.mod go.sum
	GOOS=$(shell echo $* | cut -f1 -d-) GOARCH=$(shell echo $* | cut -f2 -d- ) GOARM=6 CGO_ENABLED=0 \
		$(GO) build \
			-ldflags "-w -s $(COMMIT_FLAG) $(VERSION_FLAG)" \
			-o $@ \
			./cmd/$(PKG_NAME)

$(PREFIX)/bin/$(PKG_NAME)_%v7$(call extension,$(GOOS)): $(shell find $(PREFIX) -type f -name "*.go") go.mod go.sum
	GOOS=$(shell echo $* | cut -f1 -d-) GOARCH=$(shell echo $* | cut -f2 -d- ) GOARM=7 CGO_ENABLED=0 \
		$(GO) build \
			-ldflags "-w -s $(COMMIT_FLAG) $(VERSION_FLAG)" \
			-o $@ \
			./cmd/$(PKG_NAME)

$(PREFIX)/bin/$(PKG_NAME)_windows-%.exe: $(shell find $(PREFIX) -type f -name "*.go") go.mod go.sum
	GOOS=windows GOARCH=$* GOARM= CGO_ENABLED=0 \
		$(GO) build \
			-ldflags "-w -s $(COMMIT_FLAG) $(VERSION_FLAG)" \
			-o $@ \
			./cmd/$(PKG_NAME)

$(PREFIX)/bin/$(PKG_NAME)_%$(TARGETVARIANT)$(call extension,$(GOOS)): $(shell find $(PREFIX) -type f -name "*.go") go.mod go.sum
	GOOS=$(shell echo $* | cut -f1 -d-) GOARCH=$(shell echo $* | cut -f2 -d- ) GOARM=$(GOARM) CGO_ENABLED=0 \
		$(GO) build \
			-ldflags "-w -s $(COMMIT_FLAG) $(VERSION_FLAG)" \
			-o $@ \
			./cmd/$(PKG_NAME)

$(PREFIX)/bin/$(PKG_NAME)$(call extension,$(GOOS)): $(PREFIX)/bin/$(PKG_NAME)_$(GOOS)-$(GOARCH)$(TARGETVARIANT)$(call extension,$(GOOS))
	cp $< $@

build: $(PREFIX)/bin/$(PKG_NAME)_$(GOOS)-$(GOARCH)$(TARGETVARIANT)$(call extension,$(GOOS)) $(PREFIX)/bin/$(PKG_NAME)$(call extension,$(GOOS))

ifeq ($(OS),Windows_NT)
test:
	$(GO) test -coverprofile=c.out ./...
else
test:
	$(GO) test -race -coverprofile=c.out ./...
endif

ifeq ($(OS),Windows_NT)
integration: $(PREFIX)/bin/$(PKG_NAME)$(call extension,$(GOOS))
	$(GO) test -v \
		-ldflags "-X `$(GO) list ./internal/tests/integration`.GomplateBinPath=$(shell cygpath -ma .)/$<" \
		./internal/tests/integration
else
integration: $(PREFIX)/bin/$(PKG_NAME)
	$(GO) test -v \
		-ldflags "-X `$(GO) list ./internal/tests/integration`.GomplateBinPath=$(shell pwd)/$<" \
		./internal/tests/integration
endif

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
	@golangci-lint run --verbose --max-same-issues=0 --max-issues-per-linter=0

ci-lint:
	@golangci-lint run --verbose --max-same-issues=0 --max-issues-per-linter=0 --out-format=github-actions

.PHONY: gen-changelog clean test build-x compress-all build-release build test-integration-docker gen-docs lint clean-images clean-containers docker-images
.DELETE_ON_ERROR:
.SECONDARY:
