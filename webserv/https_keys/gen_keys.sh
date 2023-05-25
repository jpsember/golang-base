#!/usr/bin/env bash
set -eu

openssl req -x509 -out animalaid.org.crt -keyout animalaid.org.key \
  -newkey rsa:2048 -nodes -sha256 \
  -subj '/CN=animalaid.org' -extensions EXT -config <( \
   printf "[dn]\nCN=animalaid.org\n[req]\ndistinguished_name = dn\n[EXT]\nsubjectAltName=DNS:animalaid.org\nkeyUsage=digitalSignature\nextendedKeyUsage=serverAuth")
