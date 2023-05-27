#!/usr/bin/env bash
set -eu

datagen language go source_path gen clean "$@"
