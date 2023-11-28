#!/bin/bash -l

module load slurm

export START_FORECAST=2023-03-21-00 
export END_FORECAST=2023-03-21-12 
export FORECAST_DURATION=48

source ../scripts/options.sh

salloc $SALLOC_OPT --nodes 8 ../scripts/runwrfita.sh
