#!/usr/bin/env bash
set -eu

echo "Installing various dependencies"

sudo apt-get update
sudo snap install go --classic

echo "Creating a database directory"

mkdir -p sqlite

