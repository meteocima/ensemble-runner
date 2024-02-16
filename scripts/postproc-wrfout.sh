#!/bin/bash
set -e
sleep 30
mkdir -p $SIM_WORKDIR/results/out
regridded=$SIM_WORKDIR/results/out/out_regr_${INSTANT}.grb

wrk_dir=$SIM_WORKDIR/upp_wd/${INSTANT}
mkdir -vp $wrk_dir
cd $wrk_dir

dirprep $ROOTDIR/templates/upp $wrk_dir

export tmmark=d03
export MP_SHARED_MEMORY=no
export MP_LABELIO=no

mpirun -n 1 ./unipost.exe

mv -v WRFPRS* $regridded