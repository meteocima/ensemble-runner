#!/bin/bash
set -e
cd $SIM_WORKDIR/results/aux;

RH_EXPR="RH2=100*(PSFC*Q2/0.622)/(611.2*exp(17.67*(T2-273.15)/((T2-273.15)+243.5)))"

###########################
# Domain 3
###########################

# Merge all files into one that contains all simulation hours
cdo -O -v mergetime aux-regr-d03-*.nc raw-d03-${START_FORECAST}.nc

# Add RH variable
cdo -O -v -setrtoc,100,1.e99,100 -setunit,"%" -expr,$RH_EXPR raw-d03-${START_FORECAST}.nc rh-d03-${START_FORECAST}.nc

# Merge source file and RH file
cdo -O -v -z zip_2 merge raw-d03-${START_FORECAST}.nc rh-d03-${START_FORECAST}.nc regr-d03-${START_FORECAST}.nc

rm raw-d03-${START_FORECAST}.nc rh-d03-${START_FORECAST}.nc

###########################
# Domain 1
###########################

# Merge all files into one that contains all simulation hours
cdo -O -v mergetime aux-regr-d01-*.nc regr-d01-${START_FORECAST}.nc

