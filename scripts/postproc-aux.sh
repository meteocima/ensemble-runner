#!/bin/bash
set -e

src=$1
domain=$2

novar=../regridded_aux/novar_`basename $src`
regridded=../regridded_aux/regr_`basename $src`
mkdir -p `dirname $regridded`

#ncks -O -x -v P_PL,C1H,C2H,C3H,C4H,C1F,C2F,C3F,C4F $src $novar
cp $src $novar

script="$WRFITA_ROOTDIR/scripts/grid_wrfita-d0$domain.txt"
cdo_params="setreftime,2000-01-01,00:00:00 -setcalendar,standard -remapbil,$script -selgrid,1 "
cdo $cdo_params $novar $regridded
rm $novar
