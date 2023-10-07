#!/usr/bin/env bash
set -eu

# Add -s for silent
#curl  --location https://github.com/validator/validator/releases/download/20.6.30/vnu.jar_20.6.30.zip --output validator.zip
unzip -p validator.zip dist/vnu.jar > vnu.jar

#-output css-validator.jar
