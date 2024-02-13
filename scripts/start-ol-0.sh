#!/bin/bash
export PRG=$WORK/prg
export DEPS=$PRG/deps/out
export ROOTDIR=$WORK/ol

PATH=$PATH:$ROOTDIR/bin:$ROOTDIR/scripts:$PRG
export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:$DEPS/lib

module use /leonardo/prod/opt/modulefiles/global/libraries
module use /leonardo/prod/opt/modulefiles/global/compilers
module use /leonardo/prod/opt/modulefiles/global/tools
module load intel-oneapi-mpi/2021.10.0
module load intel-oneapi-compilers/2023.2.1

export DURATION_HOURS=48

gfsdn -c $PRG/gfs.toml -o $WORK/gfs  ol $DURATION_HOURS `date '+%Y%m%d00'`
sbatch $ROOTDIR/scripts/start-ol-1.sh 
