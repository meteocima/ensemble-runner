#!/bin/bash
set -e

RH_EXPR="RH2=100*(PSFC*Q2/0.622)/(611.2*exp(17.67*(T2-273.15)/((T2-273.15)+243.5)))"


module load python/3.11.6--gcc--8.5.0

# create directories if they don't exist
mkdir -p $SIM_WORKDIR/results/aux
mkdir -p $SIM_WORKDIR/results/rawaux

# results filename
regridded=$SIM_WORKDIR/results/aux/aux-regr-d0${DOMAIN}-${INSTANT}.nc

# copy original AUX file to rawaux directory to later send to continuum
cp -v ${FILE_PATH} $SIM_WORKDIR/results/rawaux/${FILE}


# fix date and time
time=`basename ${FILE} | cut -c 26-33`
date=`basename ${FILE} | cut -c 15-24`
cdo -b F64 settaxis,$date,$time ${FILE_PATH} ${FILE_PATH}.timefix1
cdo setreftime,2000-01-01,00:00:00 -setcalendar,standard ${FILE_PATH}.timefix1 ${FILE_PATH}.timefix2

# remap to dewetra grid
script="${ROOTDIR}/scripts/cdo_wrfarpal-d0${DOMAIN}_grid.txt"
cdo -remapbil,${script} -selgrid,1  ${FILE_PATH}.timefix2 ${FILE_PATH}.remapd

# fix vertical levels
python3 ${ROOTDIR}/scripts/add_lev_variable.py

# Add RH variable
cdo -O -v -setrtoc,100,1.e99,100 -setunit,"%" -expr,$RH_EXPR ${FILE_PATH}.levfixd ${FILE_PATH}.rh

# Merge original file and RH file into final results file
cdo -O -v merge ${FILE_PATH}.levfixd ${FILE_PATH}.rh $regridded

# remove temp files
rm -v ${FILE_PATH}.timefix2 ${FILE_PATH}.timefix1 ${FILE_PATH}.remapd ${FILE_PATH}.levfixd ${FILE_PATH}.rh

