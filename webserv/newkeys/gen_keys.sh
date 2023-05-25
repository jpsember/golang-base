#!/usr/bin/env bash
set -eu

# From https://letsencrypt.org/docs/certificates-for-localhost/#making-and-trusting-your-own-certificates

# But, Boris says don't use localhost; put in something like your 'real' domain name


echo "1. run this script"

echo "Creating certificate and key"

openssl req -x509 -out animalaid.crt -keyout animalaid.key \
  -newkey rsa:2048 -nodes -sha256 \
  -subj '/CN=animalaid.org' -extensions EXT -config <( \
   printf "[dn]\nCN=animalaid.org\n[req]\ndistinguished_name = dn\n[EXT]\nsubjectAltName=DNS:animalaid.org\nkeyUsage=digitalSignature\nextendedKeyUsage=serverAuth")

# "...and install localhost.crt in your list of locally trusted roots."


echo "2. double click on the crt file so it gets opened in OSX Keychain app"

echo "3. right click on it in the keychain app, get info"

echo "4. expand the trust tab, and change 'when using this certificate' to 'always trust'"

echo "5. close the window, and it will ask you to enter your user password"

echo "6. in an editor, 'sudo vi /etc/hosts' and add (after the localhost one) this line: 127.0.0.1 animalaid.org"

echo "7. if necessary, change the code to load these keys instead"

echo "8. start the go program, with --https, and from a browser, go to url https://animalaid.org/hello"

