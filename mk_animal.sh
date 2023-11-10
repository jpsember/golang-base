#!/usr/bin/env bash
set -eu

echo "Generating data classes"
dgen.sh

echo "Compiling program"
(cd cmd; env GOOS=linux GOARCH=amd64 go build animal_demo.go)

echo "Copying to install location"
cp cmd/animal_demo remotes/linode/animal_demo

echo "Pushing to remote"
dev push remotes/linode
