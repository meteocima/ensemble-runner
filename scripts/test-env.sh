#!/usr/bin/bash
export WRF_DIR=$PWD/fixtures/testrun/WRFPrg/
export WPS_DIR=$PWD/fixtures/testrun/WPSPrg/
export WRFDA_DIR=$PWD/fixtures/testrun/WRFDAPrg/
export START_FORECAST=2020-12-25-00
export DURATION_HOURS=12
export ROOTDIR=$PWD/build/
PATH+=:$PWD/fixtures/testbin/:$PWD/build/bin/
