#export GEOG_DATA=/data/unsafe/geog
#export GFS=/data/unsafe/gfs
#export BE_DIR=/data/safe/wrfita/
#export MPIOPTS="--prefix /opt/share/comps/gcc-8.3.1/ompi-4.1.4 -bind-to core -x LD_LIBRARY_PATH"
#export SALLOC_OPT="--partition=wres --ntasks-per-node=48 --ntasks-per-core=1 --ntasks-per-socket=24 --threads-per-core=1"
#export OMPI_MCA_btl=^openib
#export WRFITA_ROOTDIR=/data/safe/wrfita
#export OB_DATDIR=/data/safe/nowcasting/obs/

# export OMPI_MCA_opal_common_ucx_opal_mem_hooks=1
# export OMPI_MCA_pml_ucx_verbose=100


export SALLOC_OPT="--partition=long --ntasks-per-node=48 --ntasks-per-core=1 --ntasks-per-socket=24 --threads-per-core=1"
export OMPI_MCA_btl=^openib
export WRFITA_ROOTDIR=/home/andrea.parodi/repos/wrfita/build
