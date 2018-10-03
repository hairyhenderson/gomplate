FROM golang:1.11.1-alpine@sha256:5c9fd3699dafa79105601a8a22e94ea0a3e1d0ce083ac8827f85baa9f0a530ae AS build

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

CMD [ "--help" ]

FROM alpine:3.8@sha256:7043076348bf5040220df6ad703798fd8593a0918d06d3ce30c6c93be117e430 AS gomplate-alpine

ARG BUILD_DATE
ARG VCS_REF
ARG OS=linux
ARG ARCH=amd64

LABEL org.opencontainers.image.created=$BUILD_DATE \
      org.opencontainers.image.revision=$VCS_REF \
      org.opencontainers.image.source="https://github.com/hairyhenderson/gomplate"

RUN apk add --no-cache ca-certificates
COPY --from=artifacts /bin/gomplate_${OS}-${ARCH}-slim /bin/gomplate

ENTRYPOINT [ "/gomplate" ]

CMD [ "--help" ]

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

CMD [ "--help" ]
