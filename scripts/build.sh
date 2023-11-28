#!/bin/bash

set -e 

rm -rf ./build
mkdir -vp ./build/bin
go build -o ./build/bin/chdates ./cli/chdates
go build -o ./build/bin/dryrun ./cli/dryrun
go build -o ./build/bin/prepvars ./cli/prepvars
go build -o ./build/bin/wrfita ./cli/wrfita
go build -o ./build/bin/postproccer ./cli/postproccer
cd ../dirprep
go build -o ../wrfita/build/bin/dirprep ./cli/dirprep
cd ../wrfstats
go build -o ../wrfita/build/bin/wrfstats ./cli/wrfstats
go build -o ../wrfita/build/bin/tables ./cli/tables
cd ../wrfita2024
cp -v config.yaml ./build
cp -rv templates ./build
cp -rv scripts ./build
mkdir -vp ./build/workdir
ln -vs /data/safe/wrfita ./build/be
ln -vs /data/safe/nowcasting/obs ./build/obs