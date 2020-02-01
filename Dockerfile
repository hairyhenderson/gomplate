FROM hairyhenderson/upx:3.94 AS upx

FROM golang:1.13.7-alpine3.11 AS build

RUN apk add --no-cache \
    make \
    libgcc libstdc++ ucl \
    git

COPY --from=upx /usr/bin/upx /usr/bin/upx

RUN mkdir -p /go/src/github.com/hairyhenderson/gomplate
WORKDIR /go/src/github.com/hairyhenderson/gomplate
COPY . /go/src/github.com/hairyhenderson/gomplate

RUN make build compress

FROM scratch AS artifacts

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /go/src/github.com/hairyhenderson/gomplate/bin/* /bin/

CMD [ "/bin/gomplate_linux-amd64" ]

FROM scratch AS gomplate

ARG VCS_REF
ARG OS=linux
ARG ARCH=amd64

LABEL org.opencontainers.image.revision=$VCS_REF \
      org.opencontainers.image.source="https://github.com/hairyhenderson/gomplate"

COPY --from=artifacts /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=artifacts /bin/gomplate_${OS}-${ARCH} /gomplate

ENTRYPOINT [ "/gomplate" ]

FROM alpine:3.11.3 AS gomplate-alpine

ARG VCS_REF
ARG OS=linux
ARG ARCH=amd64

LABEL org.opencontainers.image.revision=$VCS_REF \
      org.opencontainers.image.source="https://github.com/hairyhenderson/gomplate"

RUN apk add --no-cache ca-certificates
COPY --from=artifacts /bin/gomplate_${OS}-${ARCH}-slim /bin/gomplate

ENTRYPOINT [ "/bin/gomplate" ]

FROM scratch AS gomplate-slim

ARG VCS_REF
ARG OS=linux
ARG ARCH=amd64

LABEL org.opencontainers.image.revision=$VCS_REF \
      org.opencontainers.image.source="https://github.com/hairyhenderson/gomplate"

COPY --from=artifacts /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=artifacts /bin/gomplate_${OS}-${ARCH}-slim /gomplate

ENTRYPOINT [ "/gomplate" ]
