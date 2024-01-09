#!/bin/bash
set -e

RH_EXPR="RH2=100*(PSFC*Q2/0.622)/(611.2*exp(17.67*(T2-273.15)/((T2-273.15)+243.5)))"
PATH+=:$CDO/bin
LD_LIBRARY_PATH+=:$CDO/lib

function regrid_date() {
	SRC_DIR=$1
	ENS_NUM=$2
	cd $SRC_DIR;
	echo REGRIDDING $SRC_DIR;

	if [ `ls -1 auxhist23_d03_* 2>/dev/null | wc -l ` -gt 0 ]; then
		echo	    
	else
	    echo ERROR: no aux files found for date $START_FORECAST in directory $SRC_DIR
	    exit 1
	fi

	rm *.nc || echo no previous regridded files found

	auxfiles=`ls -fd auxhist23_d03_*`
	
	for auxf in $auxfiles; do
		echo regridding $auxf
		cdo setreftime,2000-01-01,00:00:00 -setcalendar,standard -remapbil,$ROOTDIR/scripts/italy-cdo-d03-grid.txt -selgrid,1 $auxf regrid-$auxf.nc
	done

	# Merge all files into one that contains all simulation hours
	cdo -v mergetime regrid* raw-$START_FORECAST.nc
	
	# Calculate RH variable
	cdo -L -setrtoc,100,1.e99,100 -setunit,"%" -expr,$RH_EXPR raw-$START_FORECAST.nc rh-$START_FORECAST.nc

	# Merge source file and RH file
	RESULT_NAME=lexis-italy-${START_FORECAST//-/}_$ENS_NUM.nc
        cdo -v merge raw-$START_FORECAST.nc rh-$START_FORECAST.nc $RESULT_NAME
	RESULT_DIR=$ROOTDIR/results/$START_FORECAST/
		
	mkdir -p $RESULT_DIR
	mv -v $RESULT_NAME $RESULT_DIR
}

regrid_date $ROOTDIR/workdir/$START_FORECAST/wrf00 0
for ensdir in `find $ROOTDIR/workdir/$START_FORECAST/ -type d -name 'wrf00.ens*'`; do
    file=`basename $ensdir`; 
    regrid_date $ensdir ${file#*.ens}    
done
			
