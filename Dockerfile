# syntax=docker/dockerfile:1@sha256:2780b5c3bab67f1f76c781860de469442999ed1a0d7992a5efdf2cffc0e3d769
FROM --platform=$BUILDPLATFORM golang:1.26-alpine@sha256:91eda9776261207ea25fd06b5b7fed8d397dd2c0a283e77f2ab6e91bfa71079d AS build

ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT
ARG VERSION
ENV GOOS=$TARGETOS GOARCH=$TARGETARCH VERSION=$VERSION

RUN apk add --no-cache make git

WORKDIR /go/src/github.com/hairyhenderson/gomplate
COPY go.mod /go/src/github.com/hairyhenderson/gomplate
COPY go.sum /go/src/github.com/hairyhenderson/gomplate

RUN --mount=type=cache,id=go-build-${TARGETOS}-${TARGETARCH}${TARGETVARIANT},target=/root/.cache/go-build \
	--mount=type=cache,id=go-pkg-${TARGETOS}-${TARGETARCH}${TARGETVARIANT},target=/go/pkg \
		go mod download -x

COPY . /go/src/github.com/hairyhenderson/gomplate

RUN --mount=type=cache,id=go-build-${TARGETOS}-${TARGETARCH}${TARGETVARIANT},target=/root/.cache/go-build \
	--mount=type=cache,id=go-pkg-${TARGETOS}-${TARGETARCH}${TARGETVARIANT},target=/go/pkg \
		make build
RUN mv bin/gomplate* /bin/

FROM scratch AS gomplate-linux

ARG VCS_REF
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

LABEL org.opencontainers.image.revision=$VCS_REF \
	org.opencontainers.image.source="https://github.com/hairyhenderson/gomplate"

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /bin/gomplate_${TARGETOS}-${TARGETARCH}${TARGETVARIANT} /gomplate

ENTRYPOINT [ "/gomplate" ]

FROM alpine:3.23@sha256:5b10f432ef3da1b8d4c7eb6c487f2f5a8f096bc91145e68878dd4a5019afde11 AS gomplate-alpine

ARG VCS_REF
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

LABEL org.opencontainers.image.revision=$VCS_REF \
	org.opencontainers.image.source="https://github.com/hairyhenderson/gomplate"

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /bin/gomplate_${TARGETOS}-${TARGETARCH}${TARGETVARIANT} /bin/gomplate

ENTRYPOINT [ "/bin/gomplate" ]

FROM --platform=windows/amd64 mcr.microsoft.com/windows/nanoserver:ltsc2022@sha256:3ca7daac9e971bd440920e408a29d46545b717775794c71561d57ac2f9354a48 AS gomplate-windows
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT
COPY --from=build /bin/gomplate_${TARGETOS}-${TARGETARCH}${TARGETVARIANT}.exe /gomplate.exe

FROM gomplate-$TARGETOS AS gomplate
