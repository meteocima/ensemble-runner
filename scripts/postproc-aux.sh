#!/bin/bash
set -e

mkdir -p results/aux
mkdir -p results/rawaux
regridded=results/aux/aux-regr-d0${DOMAIN}-${INSTANT}.nc
cp -v ${FILE_PATH} results/rawaux/${FILE}

script="${ROOTDIR}/scripts/cdo_wrfarpal-d0${DOMAIN}_grid.txt"
time=`basename ${FILE} | cut -c 26-33`
date=`basename ${FILE} | cut -c 15-24`

cdo -remapbil,${script} -selgrid,1 ${FILE_PATH} ${FILE_PATH}.remapd
cdo -b F64 settaxis,$date,$time ${FILE_PATH}.remapd ${FILE_PATH}.timefix1
cdo setreftime,2000-01-01,00:00:00 -setcalendar,standard ${FILE_PATH}.timefix1 $regridded




#################################### 
#              TEMP
#################################### 
ROOTDIR=`realpath ..`
DOMAIN=3
script="${ROOTDIR}/scripts/cdo_wrfarpal-d0${DOMAIN}_grid.txt"

for FILE in `find auxhist23_d03_*`; do
FILE_PATH=$PWD/$FILE
time=`basename ${FILE} | cut -c 26-33`
date=`basename ${FILE} | cut -c 15-24`
cdo -b F64 settaxis,$date,$time ${FILE_PATH} ${FILE_PATH}.timefix1
cdo setreftime,2000-01-01,00:00:00 -setcalendar,standard ${FILE_PATH}.timefix1 ${FILE_PATH}.timefix2
cdo -remapbil,${script} -selgrid,1  ${FILE_PATH}.timefix2 ${FILE_PATH}.remapd

rm -v ${FILE_PATH}.timefix2 ${FILE_PATH}.timefix1
done
