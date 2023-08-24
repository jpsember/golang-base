#!/usr/bin/env bash
set -eu

datagen language go format source_path gen clean "$@"
datagen language go \
   format \
   source_path webapp/gen \
   dat_path webapp/dat_files \
   sql_dir webapp/db_res \
   clean "$@"
