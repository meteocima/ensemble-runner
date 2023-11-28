#!/bin/bash -l

set -e

module load slurm WRFITA

export SALLOC_OPT="--partition=long --nodefile=$WRFITA_ROOTDIR/hostfile --ntasks-per-node=48 --ntasks-per-core=1 --ntasks-per-socket=24 --threads-per-core=1"
export OMPI_MCA_btl=^openib
export WRFITA_ROOTDIR=/data/safe/wrfita2024

cd $WRFITA_ROOTDIR

salloc $SALLOC_OPT --nodes 8 ./bin/wrfita