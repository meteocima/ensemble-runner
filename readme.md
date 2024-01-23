# ensemble-runner

This repository contain GO source code of a series of commands which allows
to run a WRF simulation using either a GFS forecast as guiding conditions,
assimilating radars and weather forecast observations.

## Work environment preparation.

The command must run in a work directory containing a `config.yaml` config file. 
This file should be in `yaml`` format, and allows, among other things, to customize 
the path of all others external files and directories needed by the process.

The config files contains following variables:

* __GeogDataDir__					- path to a directory containing static geographic data used by geogrid.exe.
* __CovarMatrixesDir__				- path to a directory containing background errors of covariance matrices.
* __GeogridProc__ 					- number of MPI processes to use when running `geogrid.exe`
* __MetgridProc__ 					- number of MPI processes to use when running `metgrid.exe`
* __WrfProc__ 						- number of MPI processes to use when running `wrf.exe` to run the final forecast or ensemble members
* __WrfdaProc__ 					- number of MPI processes to use when running `dawrf_var.exe`
* __RealProc__ 						- number of MPI processes to use when running `real.exe`
* __WrfStepProcCount__ 				- number of MPI processes to use when running `wrf.exe` to run the intermediate steps
* __MpiOptions__					- additional arguments to pass in every invocation of `mpirun`
* __RunWPS__						- specify if boundary conditions are produced or read from an `inputs` directory
* __EnsembleMembers__				- number of members in the ensemble (excluding the control forecast)
* __EnsembleParallelism__			- how many ensemble members to run in parallel
* __AssimilateOnlyInnerDomain__		- when true, assimilation of observation data is done only for the innermost domain
* __AssimilateFirstCycle__			- when true, assimilation of observation data is done also in the first cycle
* __CoresPerNode__					- specify how many cores each node has

Additionally, some other informations are read from environment variables. Some of this variables
are already defined by other parts of the system, other ones change for every simulations, so it
does not make sense to have them in the config file.

* __START_FORECAST__	-	start of forecast to simulate, in format YYYY-MM-DD-HH. If `START_FORECAST` is omitted, the system find the date or dates to run by reading the file `inputs/arguments.txt`
* __DURATION_HOURS__	-	duration of the forecast. value is ignored when file `inputs/arguments.txt` is used.
* __SLURM_NODELIST__	-	contains hostnames of all available nodes for the simulation.
* __WRF_DIR__			-	path to compiled binaries of the WRF program.
* __WPS_DIR__			-	path to compiled binaries of the WPS program.
* __WRFDA_DIR__			-	path to compiled binaries of the WRF-DA program.
* __ROOTDIR__			-	path to the root directory of the simulation. This is the directory which contains `templates` directory, `workdir` directory, etc. 


## Command syntax

Run the command without arguments to start the simulation:

```bash
$ wrfda-run 
Usage: ensrunner [-p WPS|DA|WPSDA] [-i GFS|IFS] <workdir> <dates...>
```

#### $ROOTDIR directory

This is the path of the directory containing `config.yaml` config file. 
Directory `$ROOTDIR/workdir` will be used as starting work directory for the command while running the simulations.

At the end of the simulation, the directory `$ROOTDIR/workdir` will contains a subdirectory for each date of simulation ran, each one directory containing the complete three of intermediates data and log files used. These directory are named using a YYYY-MM-DD-HH format; 

Moreover, if WPS is run, an `$ROOTDIR/inputs` directory will be created containing a subdirectories for each date ran containing WPS results files. 
Another directory will contains all output files of the simulation: `$ROOTDIR/results`

#### Dates arguments

All arguments that follows `workdir` are interpreted as start dates of multiple simulations that are executed serially.

**N.B. When we refer to start date, we mean the date and hour of the first hour forecasted.**

The forecast by default last for 48h from the start date, but can be customized using dates arguments syntax (see below)
Guiding forecast and observations must contains date for the instants at 3 and 6 hours before the start date.

The system creates, under the specified work directory, a separate directory for each simulation date requested, named after the start date of the simulation.
Moreover, WPS outputs are copied in a `inputs` directory organize with a sub-directory for each simulation ran, again, named after the simulation start date.

##### Date arguments syntax

Each date argument can have one of the syntax's above. They must be separated by spaces.

1) **Single date** - a single date, specified in format YYYYMMDDHH
> e.g. 2020122523
	
2) **Date range** - a range of date, specified including, in format YYYYMMDDHH, the start and 
	end dates of range, separated by a dash `-` 
>	e.g. 2020122523-2020122623

Both 1 and 2 can be optionally followed by a comma and number of hours to forecast for the specified date or range of dates. 
If not specified, number of hours default to 48.
>	e.g. 2020122523,48


### Processes organization within the WPS and DA phases.	

The diagram above represent the main processes running in WPS and DA phases.

![Environments processes](media/ResponsibilityPerEnvironment.png)	