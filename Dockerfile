# syntax=docker/dockerfile:1.3.1-labs
FROM --platform=linux/amd64 golang:1.19-alpine@sha256:0ec0646e208ea58e5d29e558e39f2e59fccf39b7bda306cb53bbaff91919eca5 AS build

ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT
ENV GOOS=$TARGETOS GOARCH=$TARGETARCH

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

FROM alpine:3.16@sha256:e4cdb7d47b06ba0a062ad2a97a7d154967c8f83934594d9f2bd3efa89292996b AS gomplate-alpine

ARG VCS_REF
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

LABEL org.opencontainers.image.revision=$VCS_REF \
	org.opencontainers.image.source="https://github.com/hairyhenderson/gomplate"

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /bin/gomplate_${TARGETOS}-${TARGETARCH}${TARGETVARIANT} /bin/gomplate

ENTRYPOINT [ "/bin/gomplate" ]

FROM --platform=windows/amd64 mcr.microsoft.com/windows/nanoserver:2009@sha256:70ad3c3f156b1002a6a642d3c3b769264f9ca166f57eab62051f59c0dbe20a0f AS gomplate-windows
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT
COPY --from=build /bin/gomplate_${TARGETOS}-${TARGETARCH}${TARGETVARIANT}.exe /gomplate.exe

FROM gomplate-$TARGETOS AS gomplate
