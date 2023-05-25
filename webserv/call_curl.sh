#!/usr/bin/env bash
set -eu

curl -sL https://animalaid.org:443/hello

# Also, try piping to 'xxd' for a hex dump:
#
# | xxd
