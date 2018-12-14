FROM golang:1.11.3-alpine@sha256:c2c6c46c11319fd458a42aa3fc3b45e16bacb49e3f33f1e2a783f0122a9d8471 AS build

RUN apk add --no-cache \
    make \
    git \
    upx=3.94-r0

RUN mkdir -p /go/src/github.com/hairyhenderson/gomplate
WORKDIR /go/src/github.com/hairyhenderson/gomplate
COPY . /go/src/github.com/hairyhenderson/gomplate

RUN make build-x compress-all

FROM scratch AS artifacts

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /go/src/github.com/hairyhenderson/gomplate/bin/* /bin/

CMD [ "/bin/gomplate_linux-amd64" ]

FROM scratch AS gomplate

ARG BUILD_DATE
ARG VCS_REF
ARG OS=linux
ARG ARCH=amd64

LABEL org.opencontainers.image.created=$BUILD_DATE \
      org.opencontainers.image.revision=$VCS_REF \
      org.opencontainers.image.source="https://github.com/hairyhenderson/gomplate"

COPY --from=artifacts /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=artifacts /bin/gomplate_${OS}-${ARCH} /gomplate

ENTRYPOINT [ "/gomplate" ]

FROM alpine:3.8@sha256:621c2f39f8133acb8e64023a94dbdf0d5ca81896102b9e57c0dc184cadaf5528 AS gomplate-alpine

ARG BUILD_DATE
ARG VCS_REF
ARG OS=linux
ARG ARCH=amd64

LABEL org.opencontainers.image.created=$BUILD_DATE \
      org.opencontainers.image.revision=$VCS_REF \
      org.opencontainers.image.source="https://github.com/hairyhenderson/gomplate"

RUN apk add --no-cache ca-certificates
COPY --from=artifacts /bin/gomplate_${OS}-${ARCH}-slim /bin/gomplate

ENTRYPOINT [ "/bin/gomplate" ]

FROM scratch AS gomplate-slim

ARG BUILD_DATE
ARG VCS_REF
ARG OS=linux
ARG ARCH=amd64

LABEL org.opencontainers.image.created=$BUILD_DATE \
      org.opencontainers.image.revision=$VCS_REF \
      org.opencontainers.image.source="https://github.com/hairyhenderson/gomplate"

COPY --from=artifacts /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=artifacts /bin/gomplate_${OS}-${ARCH}-slim /gomplate

ENTRYPOINT [ "/gomplate" ]
