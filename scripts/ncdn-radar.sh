#!/bin/bash -l

set -e

module load slurm gcc-8.3.1/WRF-KIT2 > /dev/null 2> /dev/null
salloc -Q --partition=wres -n 1 mpirun /home/wrfprod/bin/ncdn-radar-2 $1
