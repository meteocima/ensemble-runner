#!/bin/bash
set -e

mkdir -p results/aux
mkdir -p results/rawaux
regridded=results/aux/aux-regr-d0${DOMAIN}-${INSTANT}.nc
cp -v ${FILE_PATH} results/rawaux/aux-d0${DOMAIN}-${INSTANT}.nc
script="${ROOTDIR}/scripts/cdo_wrfarpal-d0${DOMAIN}_grid.txt"

cdo_params="setreftime,2000-01-01,00:00:00 -setcalendar,standard -remapbil,${script} -selgrid,1 "
cdo ${cdo_params} ${FILE_PATH} ${regridded}

