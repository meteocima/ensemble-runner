#!/bin/bash

set -e 

rm -rf ./build
mkdir -vp ./build/bin
go build -o ./build/bin/prepvars ./cli/prepvars
go build -o ./build/bin/ensrunner ./cli/ensrunner
go build -o ./build/bin/dirprep ./cli/dirprep
go build -o ./build/bin/hosts ./cli/hosts
cp -v config.yaml ./build
cp -rv templates ./build
cp -rv scripts ./build
mkdir -vp ./build/workdir
ln -vs /data/safe/wrfita ./build/be
ln -vs /data/safe/nowcasting/obs ./build/obs