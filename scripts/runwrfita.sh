#!/bin/bash
set -x
run-wps() {
    eval `chdates -6h $(($FORECAST_DURATION+6))h`
    eval `prepvars`
    rm -rf ./wps
    dirprep --strict ~/repos/wrfita/templates/wps/ ./wps
    cd wps

    GFS_DIR=$GFS/$START_YEAR/$START_MONTH/$START_DAY/${START_HOUR}00

    mpirun $MPIOPTS -n 36 ./geogrid.exe
    ./link_grib.csh $GFS_DIR/*.grb
    ./ungrib.exe 
    ./avg_tsfc.exe 
    mpirun $MPIOPTS -n 64 ./metgrid.exe
    cd ..
}

run-wrf-step() {
    START_DELTA_HOUR=$1
    FROM_3=$2
    FROM_1_2=$3
    eval `chdates ${START_DELTA_HOUR} 3h`
    eval `prepvars`
    rm -rf ./wrf$START_HOUR
    dirprep --strict ~/repos/wrfita/templates/wrf-step/ ./wrf$START_HOUR
    
    cd wps
    cp -v ../wrf$START_HOUR/namelist.input .
    mpirun $MPIOPTS -n 361 ./real.exe
    
    cd ../wrf$START_HOUR
    ln -vs ../wps/wrfbdy_d01 .

    if [[ $FROM_3 == "WPS" ]]; then
        ln -vs ../wps/wrfinput_d* .
    else
        ln -vs ../$FROM_3/wrfvar_output ./wrfinput_d03
        ln -vs ../$FROM_1_2/wrfvar_input_d01 ./wrfinput_d01
        ln -vs ../$FROM_1_2/wrfvar_input_d02 ./wrfinput_d02
    fi

    mpirun $MPIOPTS -n 361 ./wrf.exe
    cd ..
}

run-da() {
    START_DELTA_HOUR=$1
    FROM=$2
    eval `chdates ${START_DELTA_HOUR} 3h`
    eval `prepvars`
    rm -rf ./da$START_HOUR

    dirprep --strict ~/repos/wrfita/templates/wrfda_03/ ./da$START_HOUR
    cd ./da$START_HOUR
    
    ln -vs ../wps/wrfbdy_d01 .
    ln -vs ../$FROM/wrfvar_input_d03 ./fg
   

    mpirun $MPIOPTS -n 361 ./da_wrfvar.exe
    cd ..
}

run-wrf-forecast() {
    FROM_3=$1
    FROM_1_2=$2

    eval `chdates 0h ${FORECAST_DURATION}h`
    eval `prepvars`
    rm -rf ./wrf$START_HOUR
    dirprep --strict ~/repos/wrfita/templates/wrf-forecast/ ./wrf$START_HOUR
    
    cd wps
    cp -v ../wrf$START_HOUR/namelist.input .
    mpirun $MPIOPTS -n 361 ./real.exe

    cd ../wrf$START_HOUR
    ln -vs ../wps/wrfbdy_d01 .
    ln -vs ../$FROM_3/wrfvar_output ./wrfinput_d03
    ln -vs ../$FROM_1_2/wrfvar_input_d01 ./wrfinput_d01
    ln -vs ../$FROM_1_2/wrfvar_input_d02 ./wrfinput_d02

    mpirun $MPIOPTS -n 361 ./wrf.exe
    cd ..
}

set -e 


source ../scripts/options.sh
ml WRFITA/3.0.0-omp

echo start wps
# run-wps
echo start first cycle wrf step
#run-wrf-step -6h WPS
echo start second cycle da
#run-da -3h wrf18

echo start second cycle wrf step
#run-wrf-step -3h da21 wrf18

echo start third cycle da
#run-da 0h wrf21

echo start main forecast
run-wrf-forecast da00 wrf21