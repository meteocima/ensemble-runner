#!/bin/bash -l

module load slurm
salloc --partition=long -n 1  /data/safe/wrfita2024/scripts/schedule-1.sh $1
