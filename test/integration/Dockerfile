FROM alpine:edge

ENV VAULT_VER 0.7.0
RUN apk add --no-cache \
    curl \
    bash \
    bats \
  && curl -L -o /tmp/vault.zip https://releases.hashicorp.com/vault/${VAULT_VER}/vault_${VAULT_VER}_linux_amd64.zip \
  && unzip /tmp/vault.zip \
  && mv vault /bin/vault \
  && rm /tmp/vault.zip \
  && apk del curl

RUN mkdir /lib64 \
  && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

COPY gomplate /bin/gomplate
COPY *.sh /tests/
COPY *.bash /tests/
COPY *.bats /tests/

CMD ["/tests/test.sh"]
