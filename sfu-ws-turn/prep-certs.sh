#!/bin/bash

set -eux

CONF_DIR=$(pwd)

TMP="/tmp/sfu-ws-turn-certs/"
echo $TMP

mkdir -p $TMP
cd $TMP

openssl genrsa -out rootCA.key 2048

openssl req -x509 -new -nodes -key rootCA.key -sha256 -days 1460 -out rootCA.pem -config $CONF_DIR/ca.cnf


openssl req -new -nodes -sha256 -out server.csr -newkey rsa:2048 -keyout server.key -config $CONF_DIR/server.csr.cnf

openssl x509 -req -in server.csr -CA rootCA.pem -CAkey rootCA.key -CAcreateserial -out server.crt -days 500 -sha256 -extfile $CONF_DIR/server.csr.cnf
