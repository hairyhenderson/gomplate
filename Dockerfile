FROM golang:1.8-alpine AS build

RUN mkdir -p /go/src/github.com/hairyhenderson/gomplate
WORKDIR /go/src/github.com/hairyhenderson/gomplate
COPY . /go/src/github.com/hairyhenderson/gomplate

RUN apk add --no-cache \
    make \
    git

RUN make build

FROM alpine:3.5

ARG BUILD_DATE
ARG VCS_REF

LABEL org.label-schema.build-date=$BUILD_DATE \
      org.label-schema.vcs-ref=$VCS_REF \
      org.label-schema.vcs-url="https://github.com/hairyhenderson/gomplate"

COPY --from=build /go/src/github.com/hairyhenderson/gomplate/bin/gomplate /usr/bin/gomplate

ENTRYPOINT [ "gomplate" ]

CMD [ "--help" ]
