# syntax=docker/dockerfile:1@sha256:4edf897a3ffa55b89f906fc8cc78afdb3f1834cc9c7083565e611a8a7d5fe99e
FROM --platform=$BUILDPLATFORM golang:1.27-alpine@sha256:d200d2dd76b03a99b52651fd1325aff7dcd35422c35c01bac56571b0507f3deb AS build

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

FROM alpine:3.24@sha256:be2656061b536c1883c4034790e21d0a09c5aaaf51968b89368a3c0371ea2c9a AS gomplate-alpine

ARG VCS_REF
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

LABEL org.opencontainers.image.revision=$VCS_REF \
	org.opencontainers.image.source="https://github.com/hairyhenderson/gomplate"

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /bin/gomplate_${TARGETOS}-${TARGETARCH}${TARGETVARIANT} /bin/gomplate

ENTRYPOINT [ "/bin/gomplate" ]

FROM --platform=windows/amd64 mcr.microsoft.com/windows/nanoserver:ltsc2022@sha256:62d28cb7771b0cf1b37db2662e468dad76e7ef15ca4ef52b2ab73c102a23da98 AS gomplate-windows
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT
COPY --from=build /bin/gomplate_${TARGETOS}-${TARGETARCH}${TARGETVARIANT}.exe /gomplate.exe

FROM gomplate-$TARGETOS AS gomplate
