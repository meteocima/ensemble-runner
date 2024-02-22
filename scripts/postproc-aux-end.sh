#!/bin/bash
set -e

cd $SIM_WORKDIR/results/aux;

cdo -O -v -z zip_4 mergetime aux-regr-d03-*.nc regr-d03-${START_FORECAST}.nc
cdo -O -v -z zip_4 mergetime aux-regr-d01-*.nc regr-d01-${START_FORECAST}.nc

