#!/bin/bash
set -e

DIR=$(dirname $0)

mkdir -p /tmp/kube
mkdir -p /tmp/kube/{newcerts,certs,crl,newcerts,private}
touch /tmp/kube/index.txt
echo "1000" >/tmp/kube/serial

cp $DIR/openssl.cnf /tmp/kube/

cd /tmp/kube

openssl genrsa -out ca.key 2048
openssl req -config openssl.cnf -key ca.key -new -x509 -days 7300 -sha256 -out ca.crt
openssl genrsa -out server.key 2048
openssl req -new -sha256 -key server.key -out server.csr -config openssl.cnf  -extensions v3_req
yes|openssl ca -config openssl.cnf -extensions v3_req -days 375 -notext -md sha256 -in server.csr -out server.crt

cd -

cp /tmp/kube/server.{crt,key} /tmp/kube/ca.crt $DIR/

