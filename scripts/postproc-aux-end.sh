#!/bin/bash

if [ "$#" -ne 2 ]; then
  echo "parameters: dir runDate (YYYYMMDDHHmm)"
  exit
fi

set -e

SRC_DIR=$1
RUNDATE=$2

cd $SRC_DIR;

RH_EXPR="RH2=100*(PSFC*Q2/0.622)/(611.2*exp(17.67*(T2-273.15)/((T2-273.15)+243.5)))"
RAINSUM_EXPR="RAINSUM=RAINNC+RAINC"

RUN_HOUR=${RUNDATE:8:4}

#########################
# 		DOMAIN 02		#
#########################

# Merge all files into one that contains all simulation hours
cdo -O -v mergetime sft_rftm_rg_wrfita_aux_d02_* ../raw_rg_wrfita-${RUNDATE}_${RUN_HOUR}UTC.nc
cd ..

# Calculate RH variable
cdo -L -setrtoc,100,1.e99,100 -setunit,"%" -expr,$RH_EXPR raw_rg_wrfita-${RUNDATE}_${RUN_HOUR}UTC.nc rh_rg_wrfita-${RUNDATE}_${RUN_HOUR}UTC.nc
# Calculate RAINSUM variable
cdo -L -setrtoc,100,1.e99,100 -setunit,"mm" -expr,$RAINSUM_EXPR raw_rg_wrfita-${RUNDATE}_${RUN_HOUR}UTC.nc rainsum_rg_wrfita-${RUNDATE}_${RUN_HOUR}UTC.nc

# Merge source file and RH file
cdo -O -v -z zip_2 merge raw_rg_wrfita-${RUNDATE}_${RUN_HOUR}UTC.nc rh_rg_wrfita-${RUNDATE}_${RUN_HOUR}UTC.nc raw_rg_wrfita_rh-${RUNDATE}_${RUN_HOUR}UTC.nc
# Merge with RAINSUM file
cdo -O -v -z zip_2 merge raw_rg_wrfita_rh-${RUNDATE}_${RUN_HOUR}UTC.nc rainsum_rg_wrfita-${RUNDATE}_${RUN_HOUR}UTC.nc rg_wrfita_d02-${RUNDATE}_${RUN_HOUR}UTC.nc

DST_SERVER=wrfprod@130.251.104.19
DST_BASEDIR=/share/archivio/experience/data/MeteoModels/WRF_DA_ITA_d02
DST_PATH=${DST_BASEDIR}/${RUNDATE:0:4}/${RUNDATE:4:2}/${RUNDATE:6:2}/${RUNDATE:8:4}

##### ssh -i /home/antonio/.ssh/id_rsa.antonio $DST_SERVER mkdir -p ${DST_PATH}
##### scp -i /home/antonio/.ssh/id_rsa.antonio rg_wrfita_d02-${RUNDATE}_${RUN_HOUR}UTC.nc ${DST_SERVER}:${DST_PATH}/rg_wrfita_d02-${RUNDATE}_${RUN_HOUR}UTC.nc.tmp
##### ssh -i /home/antonio/.ssh/id_rsa.antonio $DST_SERVER mv ${DST_PATH}/rg_wrfita_d02-${RUNDATE}_${RUN_HOUR}UTC.nc.tmp ${DST_PATH}/rg_wrfita_d02-${RUNDATE}_${RUN_HOUR}UTC.nc
##### 
rm raw_rg_wrfita-${RUNDATE}_${RUN_HOUR}UTC.nc rh_rg_wrfita-${RUNDATE}_${RUN_HOUR}UTC.nc rainsum_rg_wrfita-${RUNDATE}_${RUN_HOUR}UTC.nc raw_rg_wrfita_rh-${RUNDATE}_${RUN_HOUR}UTC.nc

#########################
# 		DOMAIN 03		#
#########################

cd $SRC_DIR;

# Merge all files into one that contains all simulation hours
cdo -O -v mergetime sft_rftm_rg_wrfita_aux_d03_* ../raw_rg_wrfita-d03-${RUNDATE}_${RUN_HOUR}UTC.nc
cd ..

# Calculate RH variable
cdo -L -setrtoc,100,1.e99,100 -setunit,"%" -expr,$RH_EXPR raw_rg_wrfita-d03-${RUNDATE}_${RUN_HOUR}UTC.nc rh_rg_wrfita-d03-${RUNDATE}_${RUN_HOUR}UTC.nc

# Merge source file and RH file
cdo -O -v -z zip_2 merge raw_rg_wrfita-d03-${RUNDATE}_${RUN_HOUR}UTC.nc rh_rg_wrfita-d03-${RUNDATE}_${RUN_HOUR}UTC.nc rg_wrfita_d03-${RUNDATE}_${RUN_HOUR}UTC.nc

DST_SERVER=wrfprod@130.251.104.19
DST_BASEDIR=/share/archivio/experience/data/MeteoModels/WRF_DA_ITA_d03
DST_PATH=${DST_BASEDIR}/${RUNDATE:0:4}/${RUNDATE:4:2}/${RUNDATE:6:2}/${RUNDATE:8:4}

##### ssh -i /home/antonio/.ssh/id_rsa.antonio $DST_SERVER mkdir -p ${DST_PATH}
##### scp -i /home/antonio/.ssh/id_rsa.antonio rg_wrfita_d03-${RUNDATE}_${RUN_HOUR}UTC.nc ${DST_SERVER}:${DST_PATH}/rg_wrfita_d03-${RUNDATE}_${RUN_HOUR}UTC.nc.tmp
##### ssh -i /home/antonio/.ssh/id_rsa.antonio $DST_SERVER mv ${DST_PATH}/rg_wrfita_d03-${RUNDATE}_${RUN_HOUR}UTC.nc.tmp ${DST_PATH}/rg_wrfita_d03-${RUNDATE}_${RUN_HOUR}UTC.nc
##### 
rm raw_rg_wrfita-d03-${RUNDATE}_${RUN_HOUR}UTC.nc rh_rg_wrfita-d03-${RUNDATE}_${RUN_HOUR}UTC.nc
