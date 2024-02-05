export ROOTDIR=$WORK/runenv2
export WRF_DIR=$WORK/prg/WRF
export WPS_DIR=$WORK/prg/WPS
export WRFDA_DIR=$WORK/prg/WRF
PATH=$PATH:$ROOTDIR/bin
module use  /leonardo/prod/opt/modulefiles
START_FORECAST=2024020100 DURATION_HOURS=48
 