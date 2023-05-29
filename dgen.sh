#!/usr/bin/env bash
set -eu

datagen language go format source_path gen clean "$@"
