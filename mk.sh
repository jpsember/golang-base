#!/usr/bin/env bash
set -eu

echo
echo Building apps
echo
dgen.sh
(cd cmd; \
  for i in *_demo.go; do
    echo $i
    go build $i
  done \
)

go mod tidy

echo
echo Unit tests
echo
go test \
 ./base/... \
 ./app/... \
 ./jt/... \
 ./webserv/...

