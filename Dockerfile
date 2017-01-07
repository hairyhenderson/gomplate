FROM golang:onbuild

ARG BUILD_DATE
ARG VCS_REF

LABEL org.label-schema.build-date=$BUILD_DATE \
      org.label-schema.vcs-ref=$VCS_REF \
      org.label-schema.vcs-url="https://github.com/hairyhenderson/gomplate"

RUN mv /go/bin/app /go/bin/gomplate

ENTRYPOINT [ "gomplate" ]

CMD [ "--help" ]
