FROM golang:1.10-alpine AS build

RUN mkdir -p /go/src/github.com/hairyhenderson/gomplate
WORKDIR /go/src/github.com/hairyhenderson/gomplate
COPY . /go/src/github.com/hairyhenderson/gomplate

RUN apk add --no-cache \
    make \
    git

RUN make build

FROM debian:jessie AS compress

RUN apt-get update -qq
RUN DEBIAN_FRONTEND=noninteractive apt-get install --no-install-recommends -yq curl xz-utils ca-certificates
RUN curl -fsSL -o /tmp/upx.tar.xz https://github.com/upx/upx/releases/download/v3.94/upx-3.94-amd64_linux.tar.xz \
  && tar Jxv -C /tmp --strip-components=1 -f /tmp/upx.tar.xz

COPY --from=build /go/src/github.com/hairyhenderson/gomplate/bin/gomplate /gomplate
RUN /tmp/upx --lzma /gomplate -o /gomplate-slim

FROM alpine:3.6 AS gomplate

ARG BUILD_DATE
ARG VCS_REF

LABEL org.label-schema.build-date=$BUILD_DATE \
      org.label-schema.vcs-ref=$VCS_REF \
      org.label-schema.vcs-url="https://github.com/hairyhenderson/gomplate"

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt 
COPY --from=build /go/src/github.com/hairyhenderson/gomplate/bin/gomplate /gomplate

ENTRYPOINT [ "/gomplate" ]

CMD [ "--help" ]

FROM alpine:3.6 AS gomplate-slim

ARG BUILD_DATE
ARG VCS_REF

LABEL org.label-schema.build-date=$BUILD_DATE \
      org.label-schema.vcs-ref=$VCS_REF \
      org.label-schema.vcs-url="https://github.com/hairyhenderson/gomplate"

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt 
COPY --from=compress /gomplate-slim /gomplate

ENTRYPOINT [ "/gomplate" ]

CMD [ "--help" ]
