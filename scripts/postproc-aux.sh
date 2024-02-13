#!/bin/bash
set -e

mkdir -p results/aux
regridded=results/aux/aux_regr_${INSTANT}.nc

script="${ROOTDIR}/scripts/italy-cdo-d0${DOMAIN}-grid.txt"

cdo_params="setreftime,2000-01-01,00:00:00 -setcalendar,standard -remapbil,${script} -selgrid,1 "
cdo ${cdo_params} ${FILE_PATH} ${regridded}
