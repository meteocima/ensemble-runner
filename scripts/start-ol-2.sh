#!/bin/bash
#SBATCH --time=3:30:00
#SBATCH -A CIM21_prod_0
#SBATCH -p dcgp_usr_prod
#SBATCH --reservation meteo_chains
#SBATCH --mem=118000
#SBATCH --qos=normal
#SBATCH --nodes=1
#SBATCH --ntasks-per-node=112
#SBATCH --error wrf_%j.err      # std-error file
#SBATCH --output wrf_%j.out     # std-output fite
#SBATCH --mail-type=ALL
#SBATCH --mail-user=andrea.parodi@cimafoundation.org
#SBATCH --job-name=postproc_WRF_cima

./bin/postproc

