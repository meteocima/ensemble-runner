#!/bin/bash
set -e

set

source=$1
instant=$3

regr_dir=`realpath $(dirname $source)/../regridded_wrfout`
regridded=$regr_dir/regr_`basename $source`

wrk_dir=$regr_dir/$instant
mkdir -vp $wrk_dir
cd $wrk_dir

ln -fs ${WRF_DIR}/run/ETAMPNEW_DATA nam_micro_lookup.dat
ln -fs ${WRF_DIR}/run/ETAMPNEW_DATA.expanded_rain hires_micro_lookup.dat
ln -fs ${UPP_DIR}/parm/wrf_cntrl.parm wrf_cntrl.parm
ln -fs ${UPP_DIR}/parm/post_avblflds.xml post_avblflds.xml
ln -fs ${UPP_DIR}/parm/postcntrl.xml postcntrl.xml
ln -fs ${UPP_DIR}/parm/postxconfig-NT.txt postxconfig-NT.txt
ln -fs ${UPP_DIR}/src/lib/g2tmpl//params_grib2_tbl_new params_grib2_tbl_new
CRTMDIR=${UPP_DIR}/src/lib/crtm2/src/fix
ln -fs $CRTMDIR/EmisCoeff/Big_Endian/EmisCoeff.bin           
ln -fs $CRTMDIR/SpcCoeff/Big_Endian/imgr_g15.SpcCoeff.bin           
ln -fs $CRTMDIR/SpcCoeff/Big_Endian/imgr_g13.SpcCoeff.bin           
ln -fs $CRTMDIR/AerosolCoeff/Big_Endian/AerosolCoeff.bin     
ln -fs $CRTMDIR/CloudCoeff/Big_Endian/CloudCoeff.bin         
ln -fs $CRTMDIR/SpcCoeff/Big_Endian/imgr_g12.SpcCoeff.bin    
ln -fs $CRTMDIR/TauCoeff/ODPS/Big_Endian/imgr_g12.TauCoeff.bin    
ln -fs $CRTMDIR/SpcCoeff/Big_Endian/imgr_g11.SpcCoeff.bin    
ln -fs $CRTMDIR/TauCoeff/ODPS/Big_Endian/imgr_g11.TauCoeff.bin    
ln -fs $CRTMDIR/SpcCoeff/Big_Endian/amsre_aqua.SpcCoeff.bin  
ln -fs $CRTMDIR/TauCoeff/ODPS/Big_Endian/amsre_aqua.TauCoeff.bin  
ln -fs $CRTMDIR/SpcCoeff/Big_Endian/tmi_trmm.SpcCoeff.bin    
ln -fs $CRTMDIR/TauCoeff/ODPS/Big_Endian/tmi_trmm.TauCoeff.bin    
ln -fs $CRTMDIR/SpcCoeff/Big_Endian/ssmi_f15.SpcCoeff.bin    
ln -fs $CRTMDIR/SpcCoeff/Big_Endian/ssmi_f13.SpcCoeff.bin    
ln -fs $CRTMDIR/SpcCoeff/Big_Endian/ssmi_f14.SpcCoeff.bin    
ln -fs $CRTMDIR/SpcCoeff/Big_Endian/ssmis_f16.SpcCoeff.bin    
ln -fs $CRTMDIR/SpcCoeff/Big_Endian/ssmis_f18.SpcCoeff.bin    
ln -fs $CRTMDIR/SpcCoeff/Big_Endian/ssmis_f19.SpcCoeff.bin    
ln -fs $CRTMDIR/SpcCoeff/Big_Endian/seviri_m10.SpcCoeff.bin    
ln -fs $CRTMDIR/SpcCoeff/Big_Endian/imgr_mt2.SpcCoeff.bin
ln -fs $CRTMDIR/SpcCoeff/Big_Endian/imgr_mt1r.SpcCoeff.bin
ln -fs $CRTMDIR/SpcCoeff/Big_Endian/imgr_insat3d.SpcCoeff.bin
ln -fs $CRTMDIR/TauCoeff/ODPS/Big_Endian/ssmi_f15.TauCoeff.bin    
ln -fs $CRTMDIR/TauCoeff/ODPS/Big_Endian/seviri_m10.TauCoeff.bin    
ln -fs $CRTMDIR/TauCoeff/ODPS/Big_Endian/imgr_mt2.TauCoeff.bin
ln -fs $CRTMDIR/TauCoeff/ODPS/Big_Endian/imgr_mt1r.TauCoeff.bin
ln -fs $CRTMDIR/TauCoeff/ODPS/Big_Endian/imgr_insat3d.TauCoeff.bin
ln -fs $CRTMDIR/TauCoeff/ODPS/Big_Endian/imgr_g15.TauCoeff.bin
ln -fs $CRTMDIR/TauCoeff/ODPS/Big_Endian/imgr_g13.TauCoeff.bin
ln -fs $CRTMDIR/TauCoeff/ODPS/Big_Endian/ssmi_f13.TauCoeff.bin
ln -fs $CRTMDIR/TauCoeff/ODPS/Big_Endian/ssmi_f14.TauCoeff.bin
ln -fs $CRTMDIR/TauCoeff/ODPS/Big_Endian/ssmis_f16.TauCoeff.bin
ln -fs $CRTMDIR/TauCoeff/ODPS/Big_Endian/ssmis_f18.TauCoeff.bin
ln -fs $CRTMDIR/TauCoeff/ODPS/Big_Endian/ssmis_f19.TauCoeff.bin
ln -fs $CRTMDIR/SpcCoeff/Big_Endian/ssmis_f20.SpcCoeff.bin   
ln -fs $CRTMDIR/TauCoeff/ODPS/Big_Endian/ssmis_f20.TauCoeff.bin   
ln -fs $CRTMDIR/SpcCoeff/Big_Endian/ssmis_f17.SpcCoeff.bin   
ln -fs $CRTMDIR/TauCoeff/ODPS/Big_Endian/ssmis_f17.TauCoeff.bin   

export tmmark=d03
export MP_SHARED_MEMORY=yes
export MP_LABELIO=yes

cat > itag <<EOF
$source
netcdf
grib2
${instant:0:10}_${instant:11:2}:00:00
NCAR
EOF

ln -fs wrf_cntrl.parm fort.14
mpirun $MPIOPTS -n 16 ${UPP_DIR}/bin/unipost.exe

mv -v WRFPRS* $regridded