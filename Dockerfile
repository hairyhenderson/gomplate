# syntax=docker/dockerfile:1@sha256:4edf897a3ffa55b89f906fc8cc78afdb3f1834cc9c7083565e611a8a7d5fe99e
FROM --platform=$BUILDPLATFORM golang:1.27-alpine@sha256:f2c872bcee949b7be9c490efd02b48a8a2fbf947cdf533f1c58525ec9d8170fd AS build

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

FROM alpine:3.24@sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6 AS gomplate-alpine

ARG VCS_REF
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

LABEL org.opencontainers.image.revision=$VCS_REF \
	org.opencontainers.image.source="https://github.com/hairyhenderson/gomplate"

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /bin/gomplate_${TARGETOS}-${TARGETARCH}${TARGETVARIANT} /bin/gomplate

ENTRYPOINT [ "/bin/gomplate" ]

FROM --platform=windows/amd64 mcr.microsoft.com/windows/nanoserver:ltsc2022@sha256:32729e1cdc6718c90952375e9bbae2b2c5d38f41ceb50a693fc2383b76df28fd AS gomplate-windows
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT
COPY --from=build /bin/gomplate_${TARGETOS}-${TARGETARCH}${TARGETVARIANT}.exe /gomplate.exe

FROM gomplate-$TARGETOS AS gomplate
