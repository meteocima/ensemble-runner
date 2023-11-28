#!/bin/bash -l

set -e

module load gcc-8.3.1/WRF-KIT > /dev/null 2> /dev/null

export WEBDROPS_AUTH_URL='https://testauth.cimafoundation.org/auth/realms/webdrops/protocol/openid-connect/token'
export WEBDROPS_URL='http://webdrops.cimafoundation.org/app/'
export WEBDROPS_USER='andrea.parodi@cimafoundation.org'
export WEBDROPS_PWD='^8J*ITws38Cd4b5Cg*g%iSni!KqMPH'
export WEBDROPS_CLIENT_ID='webdrops'

PATH+=:/data/safe/wrfita/prg/ncocdo/out/bin
LD_LIBRARY_PATH+=:/data/safe/wrfita/prg/ncocdo/out/lib
cd /data/safe/nowcasting/obs
ncdn.exe `date +%Y%m%d`${1}00 RADAR
