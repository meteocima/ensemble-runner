#!/bin/bash
export ROOTDIR=$WORK/ol
export PRG=$WORK/prg
export WRF_DIR=$PRG/WRF
export WPS_DIR=$PRG/WPS
export WRFDA_DIR=$PRG/WRF
export UPP_DIR=$PRG/UPP4.1

export DEPS=$PRG/deps/out
export UPP_DEPS=$PRG/NCEPlibs

PATH=$PATH:$ROOTDIR/bin:$ROOTDIR/scripts:$PRG
export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:$DEPS/lib:$UPP_DEPS/lib

module use /leonardo/prod/opt/modulefiles/global/libraries
module use /leonardo/prod/opt/modulefiles/global/compilers
module use /leonardo/prod/opt/modulefiles/global/tools
module load intel-oneapi-mpi/2021.10.0
module load intel-oneapi-compilers/2023.2.1
module load cdo/2.1.0--gcc--11.3.0                                                   

export HDF5=$DEPS
export NETCDF=$DEPS
export WRFIO_NCD_LARGE_FILE_SUPPORT=1 
export DURATION_HOURS=48
export START_FORECAST=${START_FORECAST:-`date '+%Y-%m-%d-00'`}

gfsdn -c $PRG/gfs.toml -o $WORK/gfs  ol $DURATION_HOURS `date '+%Y%m%d00'`
sbatch $ROOTDIR/scripts/start-ol-1.sh 
sleep 60
sbatch $ROOTDIR/scripts/start-ol-2.sh 
bin/deliver
