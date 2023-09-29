#!/usr/bin/env bash
set -eu

datagen language go format source_path gen clean "$@"

datagen --exceptions language go \
   format \
   dbsim \
   source_path webapp/gen \
   dat_path webapp/dat_files \
   "$@"

datagen --exceptions  language go \
   format \
   source_path webserv/gen \
   dat_path webserv/dat_files \
   "$@"
