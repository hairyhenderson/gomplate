.DEFAULT_GOAL = build
GO := go
extension = $(patsubst windows,.exe,$(filter windows,$(1)))
PKG_NAME := gomplate
DOCKER_REPO ?= hairyhenderson/$(PKG_NAME)
PREFIX := .
DOCKER_LINUX_PLATFORMS ?= linux/amd64,linux/arm64,linux/arm/v6,linux/arm/v7,linux/ppc64le,linux/s390x
DOCKER_PLATFORMS ?= $(DOCKER_LINUX_PLATFORMS),windows/amd64
# we just load by default, as a "dry run"
BUILDX_ACTION ?= --load
TAG_LATEST ?= latest
TAG_ALPINE ?= alpine

ifeq ("$(CI)","true")
LINT_PROCS ?= 1
else
LINT_PROCS ?= $(shell nproc)
endif

COMMIT ?= `git rev-parse --short HEAD 2>/dev/null`
VERSION ?= $(shell $(GO) run ./version/gen/vgen.go)

VERSION_PATH ?= `$(GO) list ./version`
COMMIT_FLAG ?= -X $(VERSION_PATH).GitCommit=$(COMMIT)
VERSION_FLAG ?= -X $(VERSION_PATH).Version=$(VERSION)
GO_LDFLAGS ?= $(COMMIT_FLAG) $(VERSION_FLAG)

GOOS ?= $(shell $(GO) version | sed 's/^.*\ \([a-z0-9]*\)\/\([a-z0-9]*\)/\1/')
GOARCH ?= $(shell $(GO) version | sed 's/^.*\ \([a-z0-9]*\)\/\([a-z0-9]*\)/\2/')

# allow overriding CGO_ENABLED for scenarios where gomplate must be compiled with CGO enabled, such as when using boringcrypto.
CGO_ENABLED ?= 0

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
platforms := freebsd-amd64 linux-amd64 linux-386 linux-armv6 linux-armv7 linux-arm64 linux-ppc64le linux-s390x darwin-amd64 darwin-arm64 solaris-amd64 windows-amd64.exe windows-386.exe

clean:
	rm -Rf $(PREFIX)/bin/*
	rm -f $(PREFIX)/*.[ci]id

build-x: $(patsubst %,$(PREFIX)/bin/$(PKG_NAME)_%,$(platforms))

$(PREFIX)/bin/%.zip: $(PREFIX)/bin/%
	@zip -j $@ $^

$(PREFIX)/bin/$(PKG_NAME)_windows-%.zip: $(PREFIX)/bin/$(PKG_NAME)_windows-%.exe
	@zip -j $@ $^

$(PREFIX)/bin/$(PKG_NAME)_%_checksum_sha256.txt: $(PREFIX)/bin/$(PKG_NAME)_%
	@sha256sum $< > $@

$(PREFIX)/bin/$(PKG_NAME)_%_checksum_sha512.txt: $(PREFIX)/bin/$(PKG_NAME)_%
	@sha512sum $< > $@

$(PREFIX)/bin/checksums.txt: $(PREFIX)/bin/checksums_sha256.txt
	@cp $< $@

$(PREFIX)/bin/checksums_sha256.txt: \
		$(patsubst %,$(PREFIX)/bin/$(PKG_NAME)_%_checksum_sha256.txt,$(platforms))
	@cat $^ > $@

$(PREFIX)/bin/checksums_sha512.txt: \
		$(patsubst %,$(PREFIX)/bin/$(PKG_NAME)_%_checksum_sha512.txt,$(platforms))
	@cat $^ > $@

$(PREFIX)/%.signed: $(PREFIX)/%
	@keybase sign < $< > $@

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
		--platform $(DOCKER_LINUX_PLATFORMS) \
		--tag $(DOCKER_REPO):$(TAG_ALPINE) \
		--target gomplate-alpine \
		$(BUILDX_ACTION) .

%.cid: %.iid
	@docker create $(shell cat $<) > $@

build-release: artifacts.cid
	@docker cp $(shell cat $<):/bin/. bin/

docker-images: gomplate.iid

GO_FILES := $(shell find . -type f -name "*.go")

$(PREFIX)/bin/$(PKG_NAME)_%v5$(call extension,$(GOOS)): $(GO_FILES)
	GOOS=$(shell echo $* | cut -f1 -d-) GOARCH=$(shell echo $* | cut -f2 -d- ) GOARM=5 CGO_ENABLED=$(CGO_ENABLED) \
		$(GO) build \
			-ldflags "-w -s $(GO_LDFLAGS)" \
			-o $@ \
			./cmd/$(PKG_NAME)

$(PREFIX)/bin/$(PKG_NAME)_%v6$(call extension,$(GOOS)): $(GO_FILES)
	GOOS=$(shell echo $* | cut -f1 -d-) GOARCH=$(shell echo $* | cut -f2 -d- ) GOARM=6 CGO_ENABLED=$(CGO_ENABLED) \
		$(GO) build \
			-ldflags "-w -s $(GO_LDFLAGS)" \
			-o $@ \
			./cmd/$(PKG_NAME)

$(PREFIX)/bin/$(PKG_NAME)_%v7$(call extension,$(GOOS)): $(GO_FILES)
	GOOS=$(shell echo $* | cut -f1 -d-) GOARCH=$(shell echo $* | cut -f2 -d- ) GOARM=7 CGO_ENABLED=$(CGO_ENABLED) \
		$(GO) build \
			-ldflags "-w -s $(GO_LDFLAGS)" \
			-o $@ \
			./cmd/$(PKG_NAME)

$(PREFIX)/bin/$(PKG_NAME)_windows-%.exe: $(GO_FILES)
	GOOS=windows GOARCH=$* GOARM= CGO_ENABLED=$(CGO_ENABLED) \
		$(GO) build \
			-ldflags "-w -s $(GO_LDFLAGS)" \
			-o $@ \
			./cmd/$(PKG_NAME)

$(PREFIX)/bin/$(PKG_NAME)_%$(TARGETVARIANT)$(call extension,$(GOOS)): $(GO_FILES)
	GOOS=$(shell echo $* | cut -f1 -d-) GOARCH=$(shell echo $* | cut -f2 -d- ) GOARM=$(GOARM) CGO_ENABLED=$(CGO_ENABLED) \
		$(GO) build \
			-ldflags "-w -s $(GO_LDFLAGS)" \
			-o $@ \
			./cmd/$(PKG_NAME)

$(PREFIX)/bin/$(PKG_NAME)$(call extension,$(GOOS)): $(PREFIX)/bin/$(PKG_NAME)_$(GOOS)-$(GOARCH)$(TARGETVARIANT)$(call extension,$(GOOS))
	cp $< $@

build: $(PREFIX)/bin/$(PKG_NAME)_$(GOOS)-$(GOARCH)$(TARGETVARIANT)$(call extension,$(GOOS)) $(PREFIX)/bin/$(PKG_NAME)$(call extension,$(GOOS))

# test with race detector on supported platforms
# windows/amd64 is supported in theory, but in practice it requires a C compiler
race_platforms := 'linux/amd64' 'darwin/amd64' 'darwin/arm64'
ifeq (,$(findstring '$(GOOS)/$(GOARCH)',$(race_platforms)))
export CGO_ENABLED=0
test:
	$(GO) test -coverprofile=c.out ./...
else
test:
	$(GO) test -race -coverprofile=c.out ./...
endif

bench.txt: go.mod go.sum $(GO_FILES)
	$(GO) test -benchmem -run=xxx -bench . ./... | tee $@

.SECONDEXPANSION:
testbin/%.test.exe: $$(shell $$(GO) list -f '{{.Dir}}' $$(subst testbin/,,$$(subst .test.exe,,$$@)))
	@GOOS=windows GOARCH=amd64 $(GO) test -c -o $@ $<

.SECONDEXPANSION:
testbin/%.test: $$(shell $$(GO) list -f '{{.Dir}}' $$(subst testbin/,,$$(subst .test,,$$@)))
	@$(GO) test -c -o $@ $<

# this is a special target for testing a package on Windows from a non-Windows
# host. It builds the Windows test binary, then SCPs it to the Windows host, and
# runs the tests there. This depends on the GO_REMOTE_WINDOWS environment
# variable being set as 'username@host'. The Windows host must have Git Bash
# installed, or maybe MSYS2, so that a number of standard Unix tools are
# available. Git must also be configured with a username and email address. See
# the GitHub workflow config in .github/workflows/build.yml for hints.
# A recent PowerShell is also required, such as version 7.3 or later.
#
# An F: drive is expected to be available, with a tmp directory. This is used
# to make sure gomplate can deal with files on a different volume.
.SECONDEXPANSION:
testbin/%.test.exe.remote: $$(shell $$(GO) list -f '{{.Dir}}' $$(subst testbin/,,$$(subst .test.exe.remote,,$$@)))
	@echo $<
	@GOOS=windows GOARCH=amd64 $(GO) test -tags timetzdata -c -o $(PREFIX)/testbin/remote-test.exe $<
	@scp -q $(PREFIX)/testbin/remote-test.exe $(GO_REMOTE_WINDOWS):/$(shell ssh $(GO_REMOTE_WINDOWS) 'echo %TEMP%' | cut -f2 -d= | sed -e 's#\\#/#g')/
	@ssh -o 'SetEnv TMP=F:\tmp' $(GO_REMOTE_WINDOWS) '%TEMP%\remote-test.exe'

# test-remote-windows runs the above target for all packages that have tests
.SECONDEXPANSION:
test-remote-windows: $$(shell $$(GO) list -f '{{ if not (eq "" (join .TestGoFiles "")) }}testbin/{{.ImportPath}}.test.exe.remote{{end}}' ./...)

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

# uses hugo modules now
# docs/themes/hugo-theme-relearn:
# 	git clone https://github.com/McShelby/hugo-theme-relearn.git $@

gen-docs:
	cd docs/; hugo

docs/content/functions/%.md: docs-src/content/functions/%.yml docs-src/content/functions/func_doc.md.tmpl
	gomplate -d data=$< -f docs-src/content/functions/func_doc.md.tmpl -o $@

# run the above target for all files found in docs-src/content/functions/*.yml
gen-func-docs: $(shell find docs-src/content/functions -name "*.yml" | sed -e 's#docs-src#docs#' -e 's#\.yml#\.md#')

# this target doesn't usually get used - it's mostly here as a reminder to myself
# hint: make sure CLOUDCONVERT_API_KEY is set ;)
gomplate.png: gomplate.svg
	cloudconvert -f png -c density=288 $^

lint:
	@golangci-lint run --verbose --max-same-issues=0 --max-issues-per-linter=0

ci-lint:
	@golangci-lint run --verbose --max-same-issues=0 --max-issues-per-linter=0 --out-format=github-actions

.PHONY: gen-changelog clean test build-x build-release build test-integration-docker gen-docs lint clean-images clean-containers docker-images integration gen-func-docs
.DELETE_ON_ERROR:
.SECONDARY:
