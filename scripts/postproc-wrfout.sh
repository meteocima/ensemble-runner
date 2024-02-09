#!/bin/bash
set -e

mkdir -p results/out
regridded=results/out/out_regr_${INSTANT}.nc

wrk_dir=upp_wd/${INSTANT}
mkdir -vp $wrk_dir

dirprep $ROOTDIR/templates/upp $wrk_dir

export tmmark=d03
export MP_SHARED_MEMORY=yes
export MP_LABELIO=yes

cat > itag <<EOF
${FILE_PATH}
netcdf
grib2
${INSTANT}
NCAR
EOF

ln -fs wrf_cntrl.parm fort.14
mpirun -n 16 ./unipost.exe

mv -v WRFPRS* $regridded