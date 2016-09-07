FROM golang:onbuild

RUN mv /go/bin/app /go/bin/gomplate

ENTRYPOINT [ "gomplate" ]

CMD [ "--help" ]
