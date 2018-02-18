extension = $(patsubst windows,.exe,$(filter windows,$(1)))
GO := go
PKG_NAME := gomplate
PREFIX := .

COMMIT ?= `git rev-parse --short HEAD 2>/dev/null`
VERSION ?= `git describe --abbrev=0 --tags $(git rev-list --tags --max-count=1) 2>/dev/null | sed 's/v\(.*\)/\1/'`

COMMIT_FLAG := -X `go list ./version`.GitCommit=$(COMMIT)
VERSION_FLAG := -X `go list ./version`.Version=$(VERSION)

GOOS ?= `go version | sed 's/^.*\ \([a-z0-9]*\)\/\([a-z0-9]*\)/\1/'`
GOARCH ?= `go version | sed 's/^.*\ \([a-z0-9]*\)\/\([a-z0-9]*\)/\2/'`

platforms := linux-amd64 linux-386 linux-arm linux-arm64 darwin-amd64 solaris-amd64 windows-amd64.exe windows-386.exe

define gocross
	GOOS=$(1) GOARCH=$(2) CGO_ENABLED=0 \
		$(GO) build \
			-ldflags "-w -s $(COMMIT_FLAG) $(VERSION_FLAG)" \
			-o $(PREFIX)/bin/$(PKG_NAME)_$(1)-$(2)$(call extension,$(1));
endef

define gocross-tool
	GOOS=$(1) GOARCH=$(2) CGO_ENABLED=0 \
		$(GO) build \
			-ldflags "-w -s $(COMMIT_FLAG) $(VERSION_FLAG)" \
			-o $(PREFIX)/bin/$(3)_$(1)-$(2)$(call extension,$(1)) \
			$(PREFIX)/test/integration/$(3)svc;
endef

define compress
	upx --lzma $(PREFIX)/bin/$(PKG_NAME)_$(1)-$(2)$(call extension,$(1)) \
	 -o $(PREFIX)/bin/$(PKG_NAME)_$(1)-$(2)-slim$(call extension,$(1))
endef

clean:
	rm -Rf $(PREFIX)/bin/*
	rm -f $(PREFIX)/test/integration/gomplate
	rm -f $(PREFIX)/test/integration/mirror
	rm -f $(PREFIX)/test/integration/meta
	rm -f $(PREFIX)/test/integration/aws

build-x: $(patsubst %,$(PREFIX)/bin/$(PKG_NAME)_%,$(platforms))

compress-all:
	$(call compress,linux,amd64)
	$(call compress,linux,arm)
	$(call compress,windows,amd64)

compress: build
	@upx --lzma $(PREFIX)/bin/$(PKG_NAME)$(call extension,$(GOOS)) \
		-o $(PREFIX)/bin/$(PKG_NAME)-slim$(call extension,$(GOOS))

build-release: clean build-x compress-all

$(PREFIX)/bin/$(PKG_NAME)_%: $(shell find $(PREFIX) -type f -name '*.go' -not -path "$(PREFIX)/test/*")
	$(call gocross,$(shell echo $* | sed 's/\([^-]*\)-\([^.]*\).*/\1/'),$(shell echo $* | sed 's/\([^-]*\)-\([^.]*\).*/\2/'))

$(PREFIX)/bin/mirror_%: $(shell find $(PREFIX)/test/integration/mirrorsvc -type f -name '*.go')
	$(call gocross-tool,$(shell echo $* | sed 's/\([^-]*\)-\([^.]*\).*/\1/'),$(shell echo $* | sed 's/\([^-]*\)-\([^.]*\).*/\2/'),mirror)

$(PREFIX)/bin/meta_%: $(shell find $(PREFIX)/test/integration/metasvc -type f -name '*.go')
	$(call gocross-tool,$(shell echo $* | sed 's/\([^-]*\)-\([^.]*\).*/\1/'),$(shell echo $* | sed 's/\([^-]*\)-\([^.]*\).*/\2/'),meta)

$(PREFIX)/bin/aws_%: $(shell find $(PREFIX)/test/integration/awssvc -type f -name '*.go')
	$(call gocross-tool,$(shell echo $* | sed 's/\([^-]*\)-\([^.]*\).*/\1/'),$(shell echo $* | sed 's/\([^-]*\)-\([^.]*\).*/\2/'),aws)

$(PREFIX)/bin/$(PKG_NAME)$(call extension,$(GOOS)): $(shell find $(PREFIX) -type f -name '*.go' -not -path "$(PREFIX)/test/*")
	CGO_ENABLED=0 \
		$(GO) build -ldflags "-w -s $(COMMIT_FLAG) $(VERSION_FLAG)" -o $@

$(PREFIX)/bin/mirror$(call extension,$(GOOS)): $(shell find $(PREFIX)/test/integration/mirrorsvc -type f -name '*.go')
	CGO_ENABLED=0 \
		$(GO) build -ldflags "-w -s $(COMMIT_FLAG) $(VERSION_FLAG)" -o $@ $(PREFIX)/test/integration/mirrorsvc

$(PREFIX)/bin/meta$(call extension,$(GOOS)): $(shell find $(PREFIX)/test/integration/metasvc -type f -name '*.go')
	CGO_ENABLED=0 \
		$(GO) build -ldflags "-w -s $(COMMIT_FLAG) $(VERSION_FLAG)" -o $@ $(PREFIX)/test/integration/metasvc

$(PREFIX)/bin/aws$(call extension,$(GOOS)): $(shell find $(PREFIX)/test/integration/awssvc -type f -name '*.go')
	CGO_ENABLED=0 \
		$(GO) build -ldflags "-w -s $(COMMIT_FLAG) $(VERSION_FLAG)" -o $@ $(PREFIX)/test/integration/awssvc

build: $(PREFIX)/bin/$(PKG_NAME)$(call extension,$(GOOS))

build-mirror: $(PREFIX)/bin/mirror$(call extension,$(GOOS))

build-meta: $(PREFIX)/bin/meta$(call extension,$(GOOS))

build-aws: $(PREFIX)/bin/aws$(call extension,$(GOOS))

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

.PHONY: gen-changelog clean test build-x compress-all build-release build build-integration-image test-integration-docker gen-docs lint
