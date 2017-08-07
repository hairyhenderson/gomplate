FROM alpine:edge

ENV VAULT_VER 0.7.3
ENV CONSUL_VER 0.9.0
RUN apk add --no-cache \
    curl \
    bash \
    bats \
    jq \
  && curl -L -o /tmp/vault.zip https://releases.hashicorp.com/vault/${VAULT_VER}/vault_${VAULT_VER}_linux_amd64.zip \
  && unzip /tmp/vault.zip \
  && mv vault /bin/vault \
  && rm /tmp/vault.zip \
  && curl -L -o /tmp/consul.zip https://releases.hashicorp.com/consul/${CONSUL_VER}/consul_${CONSUL_VER}_linux_amd64.zip \
  && unzip /tmp/consul.zip \
  && mv consul /bin/consul \
  && rm /tmp/consul.zip

RUN mkdir /lib64 \
  && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

COPY gomplate /bin/gomplate
COPY mirror /bin/mirror
COPY meta /bin/meta
COPY aws /bin/aws
COPY *.sh /tests/
COPY *.bash /tests/
COPY *.bats /tests/
COPY *.db /test/integration/

CMD ["/tests/test.sh"]
