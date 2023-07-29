#!/usr/bin/env bash
set -eu

URL=jeff

# Delete any existing certificates with that name
#
rm -f ${URL}".org.crt" ${URL}".org.key"

echo
echo Generating certificate...
echo
openssl req -x509 -days 3650 -out ${URL}".org.crt" -keyout ${URL}".org.key" \
  -newkey rsa:2048 -nodes -sha256 \
  -subj "/CN="${URL}".org" -extensions EXT -config <( \
   printf "[dn]\nCN="${URL}".org\n[req]\ndistinguished_name = dn\n[EXT]\nsubjectAltName=DNS:"${URL}".org\nkeyUsage=digitalSignature\nextendedKeyUsage=serverAuth")

echo
echo "This should report an expiry year of >= 2033:"
echo
openssl x509 -enddate -noout -in ${URL}.org.crt
