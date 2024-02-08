export ROOTDIR=/home/wrfprod/repos/ensemble-runner/build
module          load WPS/4.1-smooth
module          load WRF/4.1.5-cima
module          load WRFDA/4.1.5-cima
module          load CDO/1.7.2
module          load NCO/5.1.3
module          load UPP/3.2

PATH=$PATH:$ROOTDIR/bin
export SLURM_NODELIST=wn[01-08]
export START_FORECAST=2024-02-06-00 
export DURATION_HOURS=48
 