#!/usr/bin/env bash
set -eu

datagen language go format source_path gen clean "$@"

datagen --exceptions language go \
   dbsim \
   format \
   source_path webapp/gen \
   dat_path webapp/dat_files \
   clean "$@"
