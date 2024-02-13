#!/bin/bash
#SBATCH --time=3:30:00
#SBATCH -A CIM21_prod_0
#SBATCH -p dcgp_usr_prod 
#SBATCH --mem=118000
#SBATCH --qos=normal
#SBATCH --nodes=14
#SBATCH --ntasks-per-node=112
#SBATCH --error wrf_%j.err      # std-error file
#SBATCH --output wrf_%j.out     # std-output fite
#SBATCH --mail-type=ALL
#SBATCH --mail-user=andrea.parodi@cimafoundation.org
#SBATCH --job-name=WRF_cima

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
export START_FORECAST=`date '+%Y-%m-%d-00'`

./bin/postproc&
./bin/ensrunner
wait
