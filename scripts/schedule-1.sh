#!/bin/bash -l

set -e

export LC_ALL=en_US.UTF-8
export START=$1
if [[ $START == 00 ]]; then
    export START_FORECAST=`date +%Y-%m-%d -d '+1 day'`-$START
else
    export START_FORECAST=`date +%Y-%m-%d`-$START
fi	
ssh $SLURM_JOB_NODELIST START_FORECAST=$START_FORECAST FORECAST_DURATION=8 /data/safe/wrfita2024/scripts/schedule-2.sh
