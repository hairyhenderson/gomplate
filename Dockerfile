# syntax=docker/dockerfile:1.1.5-experimental
FROM --platform=linux/amd64 hairyhenderson/upx:3.94 AS upx

FROM --platform=linux/amd64 golang:1.14.2-alpine3.11 AS build

ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT
ARG TARGETPLATFORM
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

FROM --platform=linux/amd64 alpine:3.11.5 AS compress

ARG TARGETOS
ARG TARGETARCH

RUN apk add --no-cache \
    make \
    libgcc libstdc++ ucl

ENV GOOS=$TARGETOS GOARCH=$TARGETARCH
WORKDIR /go/src/github.com/hairyhenderson/gomplate
COPY Makefile .
RUN mkdir bin

COPY --from=upx /usr/bin/upx /usr/bin/upx
COPY --from=build bin/* bin/

RUN make compress
RUN mv bin/gomplate* /bin/

FROM scratch AS gomplate-linux

ARG VCS_REF
ARG TARGETOS
ARG TARGETARCH

LABEL org.opencontainers.image.revision=$VCS_REF \
      org.opencontainers.image.source="https://github.com/hairyhenderson/gomplate"

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /bin/gomplate_${TARGETOS}-${TARGETARCH} /gomplate

ENTRYPOINT [ "/gomplate" ]

FROM alpine:3.11.5 AS gomplate-alpine

ARG VCS_REF
ARG TARGETOS
ARG TARGETARCH

LABEL org.opencontainers.image.revision=$VCS_REF \
      org.opencontainers.image.source="https://github.com/hairyhenderson/gomplate"

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=compress /bin/gomplate_${TARGETOS}-${TARGETARCH}-slim /gomplate

ENTRYPOINT [ "/bin/gomplate" ]

FROM scratch AS gomplate-slim-linux

ARG VCS_REF
ARG TARGETOS
ARG TARGETARCH

LABEL org.opencontainers.image.revision=$VCS_REF \
      org.opencontainers.image.source="https://github.com/hairyhenderson/gomplate"

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=compress /bin/gomplate_${TARGETOS}-${TARGETARCH}-slim /gomplate

ENTRYPOINT [ "/gomplate" ]

FROM --platform=windows/amd64 mcr.microsoft.com/windows/nanoserver:1809 AS gomplate-windows
ARG TARGETOS
ARG TARGETARCH
COPY --from=build /bin/gomplate_${TARGETOS}-${TARGETARCH}.exe /gomplate.exe

FROM --platform=windows/amd64 mcr.microsoft.com/windows/nanoserver:1809 AS gomplate-slim-windows
ARG TARGETOS
ARG TARGETARCH
COPY --from=compress /bin/gomplate_${TARGETOS}-${TARGETARCH}-slim.exe /gomplate.exe

# FROM scratch AS gomplate-slim-darwin
# COPY --from=build /bin/gomplate_darwin-amd64-slim /gomplate

FROM gomplate-$TARGETOS AS gomplate
FROM gomplate-slim-$TARGETOS AS gomplate-slim
