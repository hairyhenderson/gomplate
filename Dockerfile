FROM golang:1.10-alpine@sha256:3fb74d9a61403acd5f2ce02cdd6570467350b23fcdd191628d4fddaa6523be02 AS build

RUN apk add --no-cache \
    make \
    git \
    upx

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

LABEL org.label-schema.build-date=$BUILD_DATE \
      org.label-schema.vcs-ref=$VCS_REF \
      org.label-schema.vcs-url="https://github.com/hairyhenderson/gomplate"

COPY --from=artifacts /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=artifacts /bin/gomplate_${OS}-${ARCH} /gomplate

ENTRYPOINT [ "/gomplate" ]

CMD [ "--help" ]

FROM scratch AS gomplate-slim

ARG BUILD_DATE
ARG VCS_REF
ARG OS=linux
ARG ARCH=amd64

LABEL org.label-schema.build-date=$BUILD_DATE \
      org.label-schema.vcs-ref=$VCS_REF \
      org.label-schema.vcs-url="https://github.com/hairyhenderson/gomplate"

COPY --from=artifacts /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=artifacts /bin/gomplate_${OS}-${ARCH}-slim /gomplate

ENTRYPOINT [ "/gomplate" ]

CMD [ "--help" ]
