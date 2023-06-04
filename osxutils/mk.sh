#!/usr/bin/env bash
set -eu

(cd ..; dgen.sh)
go build
