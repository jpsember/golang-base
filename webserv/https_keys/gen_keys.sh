#!/usr/bin/env bash
set -eu

URL=zebra

openssl req -x509 -out ${URL}.org.crt -keyout ${URL}.org.key \
  -newkey rsa:2048 -nodes -sha256 \
  -subj '/CN=${URL}.org' -extensions EXT -config <( \
   printf "[dn]\nCN=${URL}.org\n[req]\ndistinguished_name = dn\n[EXT]\nsubjectAltName=DNS:${URL}.org\nkeyUsage=digitalSignature\nextendedKeyUsage=serverAuth")
