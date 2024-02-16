#!/bin/bash

set -e 
TYPE=$1
rm -rf ./build
mkdir -vp ./build/bin
go build -o ./build/bin/deliver ./cli/deliver
go build -o ./build/bin/prepvars ./cli/prepvars
go build -o ./build/bin/ensrunner ./cli/ensrunner
go build -o ./build/bin/dirprep ./cli/dirprep
go build -o ./build/bin/hosts ./cli/hosts
go build -o ./build/bin/postproc ./cli/postproc
cp -v $TYPE.config.yaml ./build/config.yaml
cp -rv templates/$TYPE ./build/templates
cp -rv scripts ./build
mkdir -vp ./build/workdir
