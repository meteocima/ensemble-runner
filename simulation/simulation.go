package simulation

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/meteocima/ensemble-runner/conf"
	"github.com/meteocima/ensemble-runner/errors"
	"github.com/meteocima/ensemble-runner/folders"
	"github.com/meteocima/ensemble-runner/log"
	"github.com/meteocima/ensemble-runner/mpiman"
	"github.com/meteocima/ensemble-runner/par"
	"github.com/meteocima/ensemble-runner/server"
)

type Simulation struct {
	Start    time.Time
	Duration time.Duration
	Workdir  string
	Nodes    mpiman.SlurmNodes
}

var ShortDtFormat = "2006-01-02-15"

var join = filepath.Join

type SimDirs struct {
	wpsdir        string
	wrf18dir      string
	wrf21dir      string
	wrf00dir      string
	wpsOutputsDir string
	da18dir       []string
	da21dir       []string
	da00dir       []string
}

func (s *Simulation) run() {
	// define directories vars for the workdir of various steps of the simulation:
	// da dirs have one dir for every domain
	dirs := simDirs(s)

	// start simulation
	log.Info("Starting simulation from %s for %.0f hours", s.Start.Format(ShortDtFormat), s.Duration.Hours())
	log.Info("  -- $WORKDIR=%s", s.Workdir)

	// in case the workdir for this particular date already exists, it is removed.
	if server.DirExists(s.Workdir) {
		server.Rmdir(s.Workdir)
	}

	defer errors.OnFailuresDo(func(err errors.RunTimeError) {
		panic(err)
	})

	// assimilation happens in domain from firstDomain to 3,
	// so if the configuration specifies to assimilate only the
	// inner domain we set firstDomain=3
	firstDomain := 1
	if conf.Values.AssimilateOnlyInnerDomain {
		firstDomain = 3
	}

	// create all directories for the various wrf and wrfda cycles.
	// and, if needed, for WPS
	s.createSimulationDirectories()

	// if an ensemble is requested, create the directories for the ensemble members
	// and calculate the seed for each member
	if conf.Values.EnsembleMembers > 0 {
		rnd := rand.NewSource(0xfeedbabebadcafe)

		for ensnum := 1; ensnum <= conf.Values.EnsembleMembers; ensnum++ {
			seed := rnd.Int63()%100 + 1
			os.Setenv("ENSEMBLE_SEED", fmt.Sprintf("%02d", seed))
			log.Debug("Using seed %02d for member n.%d.", seed, ensnum)

			s.createWrfEnsembleMemberDir(s.Start, s.Duration, ensnum)
		}
	}

	// if WPS execution is requested, initial and boundary conditions are copied from the outputs of WPS.
	// otherwise, they are copied from the inputs directory.
	if conf.Values.RunWPS {
		// WPS: run geogrid, ungrib, metgrid
		// if WPS preproccing is requested in configuration
		var start time.Time
		var duration time.Duration

		// if assimilation is requested, we need to run WPS
		// for 6 hours before the start of the forecast
		if conf.Values.AssimilateObservations {
			start = s.Start.Add(-6 * time.Hour)
			duration = s.Duration + 6*time.Hour
		} else {
			start = s.Start
			duration = s.Duration
		}

		s.RunGeogrid()

		s.RunLinkGrib(start)
		s.RunUngrib()

		// if the forecast is longer than 24 hours,
		// eventually accounting for the 6 hours of
		// assimilation, we need to run avgtsfc
		if duration > 24*time.Hour {
			s.RunAvgtsfc()
		}

		s.RunMetgrid()

		// creates the directory for WPS outputs.
		// it will be called inputs because it
		// contains the input datasets for the forecast
		server.MkdirAll(dirs.wpsOutputsDir, 0775)

		if conf.Values.AssimilateObservations {
			// run real for every cycle of assimilation and for the main forecast.
			// initial conditions are copied from the outputs of execution of real.exe for the first cycle,
			// boundary conditions are copied from the outputs of execution of real.exe for every cycle.
			// namelist for the execution of the various real.exe are copied from the wrf directories of every cycle.
			server.CopyFile(s.Workdir, join(dirs.wrf18dir, "namelist.input"), join(dirs.wpsdir, "namelist.input"))
			s.RunReal(s.Start.Add(-6 * time.Hour))
			server.CopyFile(s.Workdir, join(dirs.wpsdir, "wrfbdy_d01"), join(dirs.wpsOutputsDir, "wrfbdy_d01_da01"))
			server.CopyFile(s.Workdir, join(dirs.wpsdir, "wrfinput_d01"), join(dirs.wpsOutputsDir, "wrfinput_d01"))
			server.CopyFile(s.Workdir, join(dirs.wpsdir, "wrfinput_d02"), join(dirs.wpsOutputsDir, "wrfinput_d02"))
			server.CopyFile(s.Workdir, join(dirs.wpsdir, "wrfinput_d03"), join(dirs.wpsOutputsDir, "wrfinput_d03"))

			server.CopyFile(s.Workdir, join(dirs.wrf21dir, "namelist.input"), join(dirs.wpsdir, "namelist.input"))
			s.RunReal(s.Start.Add(-3 * time.Hour))
			server.CopyFile(s.Workdir, join(dirs.wpsdir, "wrfbdy_d01"), join(dirs.wpsOutputsDir, "wrfbdy_d01_da02"))

			server.CopyFile(s.Workdir, join(dirs.wrf00dir, "namelist.input"), join(dirs.wpsdir, "namelist.input"))
			s.RunReal(s.Start)
			server.CopyFile(s.Workdir, join(dirs.wpsdir, "wrfbdy_d01"), join(dirs.wpsOutputsDir, "wrfbdy_d01_da03"))
		} else {
			// run real for the main forecast.
			// namelist for the execution of real.exe is copied from the wrf00 directory.
			// initial and boundary conditions are copied from the outputs of execution of real.exe
			server.CopyFile(s.Workdir, join(dirs.wrf00dir, "namelist.input"), join(dirs.wpsdir, "namelist.input"))
			s.RunReal(s.Start)
			server.CopyFile(s.Workdir, join(dirs.wpsdir, "wrfbdy_d01"), join(dirs.wpsOutputsDir, "wrfbdy_d01"))
			server.CopyFile(s.Workdir, join(dirs.wpsdir, "wrfinput_d01"), join(dirs.wpsOutputsDir, "wrfinput_d01"))
			server.CopyFile(s.Workdir, join(dirs.wpsdir, "wrfinput_d02"), join(dirs.wpsOutputsDir, "wrfinput_d02"))
			server.CopyFile(s.Workdir, join(dirs.wpsdir, "wrfinput_d03"), join(dirs.wpsOutputsDir, "wrfinput_d03"))
		}

	}

	// if assimilation is requested, we need to run the 3 cycles of assimilation.
	// for every cycle, we need to run WRF from the start of the cycle to the end of the cycle,
	// in order to advance the time of the initiali conditions for the next phase.
	// Result sof the last cycle will be copied to the wrf directory of the main forecast.

	// here we assimilate the 1 cycle.
	if conf.Values.AssimilateObservations {
		// Assimilation of first cycle is optional: if not requested,
		// initial and boundary conditions are copied directly from wps
		// into the first cycle wrf directory.
		if conf.Values.AssimilateFirstCycle {
			// assimilate D-6 (first cycle) for all domains
			if !conf.Values.AssimilateOnlyInnerDomain {
				server.CopyFile(s.Workdir, join(dirs.wpsOutputsDir, "wrfbdy_d01_da01"), join(dirs.da18dir[1], "wrfbdy_d01"))
			}
			for domain := firstDomain; domain <= 3; domain++ {
				// input condition are copied from wps
				server.CopyFile(s.Workdir, join(dirs.wpsOutputsDir, fmt.Sprintf("wrfinput_d%02d", domain)), join(dirs.da18dir[domain], "fg"))
				s.RunDa(s.Start.Add(-6*time.Hour), domain)
			}

			// to run WRF from D-6 to D-3, input and boundary conditions are copied from da dirs of first cycle.
			if !conf.Values.AssimilateOnlyInnerDomain {
				server.CopyFile(s.Workdir, join(dirs.da18dir[1], "wrfbdy_d01"), join(dirs.wrf18dir, "wrfbdy_d01"))
			} else {
				server.CopyFile(s.Workdir, join(dirs.wpsOutputsDir, "wrfbdy_d01_da01"), join(dirs.wrf18dir, "wrfbdy_d01"))
			}
			for domain := firstDomain; domain <= 3; domain++ {
				server.CopyFile(s.Workdir, join(dirs.da18dir[domain], "wrfvar_output"), join(dirs.wrf18dir, fmt.Sprintf("wrfinput_d%02d", domain)))
			}
			if conf.Values.AssimilateOnlyInnerDomain {
				for domain := 1; domain <= 2; domain++ {
					server.CopyFile(s.Workdir, join(dirs.wpsOutputsDir, fmt.Sprintf("wrfinput_d%02d", domain)), join(dirs.wrf18dir, fmt.Sprintf("wrfinput_d%02d", domain)))
				}
			}

		} else {
			// to run WRF from D-6 to D-3, input and boundary conditions are copied from wps.
			server.CopyFile(s.Workdir, join(dirs.wpsOutputsDir, "wrfbdy_d01_da01"), join(dirs.wrf18dir, "wrfbdy_d01"))
			for domain := 1; domain <= 3; domain++ {
				server.CopyFile(s.Workdir, join(dirs.wpsOutputsDir, fmt.Sprintf("wrfinput_d%02d", domain)), join(dirs.wrf18dir, fmt.Sprintf("wrfinput_d%02d", domain)))
			}
		}
	} else {
		server.CopyFile(s.Workdir, join(dirs.wpsOutputsDir, "wrfbdy_d01"), join(dirs.wrf00dir, "wrfbdy_d01"))
		for domain := 1; domain <= 3; domain++ {
			server.CopyFile(s.Workdir, join(dirs.wpsOutputsDir, fmt.Sprintf("wrfinput_d%02d", domain)), join(dirs.wrf00dir, fmt.Sprintf("wrfinput_d%02d", domain)))
		}
	}

	// Assimilation cycles 2 and 3
	if conf.Values.AssimilateObservations {
		// run WRF from D-6 to D-3.
		s.RunWrfStep(s.Start.Add(-6 * time.Hour))

		// assimilate D-3 (second cycle) for all domains
		if !conf.Values.AssimilateOnlyInnerDomain {
			server.CopyFile(s.Workdir, join(dirs.wpsOutputsDir, "wrfbdy_d01_da02"), join(dirs.da21dir[1], "wrfbdy_d01"))
		}
		for domain := firstDomain; domain <= 3; domain++ {
			// input condition are copied from previous cycle wrf
			server.CopyFile(s.Workdir, join(dirs.wrf18dir, fmt.Sprintf("wrfvar_input_d%02d", domain)), join(dirs.da21dir[domain], "fg"))
			s.RunDa(s.Start.Add(-3*time.Hour), domain)
		}

		// run WRF from D-3 to D. input and boundary conditions are copied from da dirs of second cycle.
		// boundary condition are copied from wps
		if conf.Values.AssimilateOnlyInnerDomain {
			server.CopyFile(s.Workdir, join(dirs.wpsOutputsDir, "wrfbdy_d01_da02"), join(dirs.wrf21dir, "wrfbdy_d01"))
		} else {
			server.CopyFile(s.Workdir, join(dirs.da21dir[1], "wrfbdy_d01"), join(dirs.wrf21dir, "wrfbdy_d01"))
		}

		for domain := firstDomain; domain <= 3; domain++ {
			server.CopyFile(s.Workdir, join(dirs.da21dir[domain], "wrfvar_output"), join(dirs.wrf21dir, fmt.Sprintf("wrfinput_d%02d", domain)))
		}
		if conf.Values.AssimilateOnlyInnerDomain {
			for domain := 1; domain <= 2; domain++ {
				server.CopyFile(s.Workdir, join(dirs.wrf18dir, fmt.Sprintf("wrfvar_input_d%02d", domain)), join(dirs.wrf21dir, fmt.Sprintf("wrfinput_d%02d", domain)))
			}
		}

		s.RunWrfStep(s.Start.Add(-3 * time.Hour))

		// assimilate D (third cycle) for all domains
		if !conf.Values.AssimilateOnlyInnerDomain {
			server.CopyFile(s.Workdir, join(dirs.wpsOutputsDir, "wrfbdy_d01_da03"), join(dirs.da00dir[1], "wrfbdy_d01"))
		}
		for domain := firstDomain; domain <= 3; domain++ {
			// input condition are copied from previous cycle wrf
			server.CopyFile(s.Workdir, join(dirs.wrf21dir, fmt.Sprintf("wrfvar_input_d%02d", domain)), join(dirs.da00dir[domain], "fg"))
			s.RunDa(s.Start, domain)
		}

		if conf.Values.AssimilateOnlyInnerDomain {
			server.CopyFile(s.Workdir, join(dirs.wpsOutputsDir, "wrfbdy_d01_da03"), join(dirs.wrf00dir, "wrfbdy_d01"))
		} else {
			server.CopyFile(s.Workdir, join(dirs.da00dir[1], "wrfbdy_d01"), join(dirs.wrf00dir, "wrfbdy_d01"))
		}
		// run WRF from D for the duration of the forecast. input and boundary conditions are copied from da dirs of third cycle.
		for domain := firstDomain; domain <= 3; domain++ {
			server.CopyFile(s.Workdir, join(dirs.da00dir[domain], "wrfvar_output"), join(dirs.wrf00dir, fmt.Sprintf("wrfinput_d%02d", domain)))
		}
		if conf.Values.AssimilateOnlyInnerDomain {
			for domain := 1; domain <= 2; domain++ {
				server.CopyFile(s.Workdir, join(dirs.wrf21dir, fmt.Sprintf("wrfvar_input_d%02d", domain)), join(dirs.wrf00dir, fmt.Sprintf("wrfinput_d%02d", domain)))
			}
		}
	}

	// if an ensemble is procduced, copy wrfinput and wrfbdy from control forecast to all ensemble members
	for ensnum := 1; ensnum <= conf.Values.EnsembleMembers; ensnum++ {
		ensdir := folders.WrfEnsembleProcWorkdir(s.Workdir, s.Start, ensnum)
		server.CopyFile(s.Workdir, join(dirs.wrf00dir, "wrfinput_d01"), join(ensdir, "wrfinput_d01"))
		server.CopyFile(s.Workdir, join(dirs.wrf00dir, "wrfinput_d02"), join(ensdir, "wrfinput_d02"))
		server.CopyFile(s.Workdir, join(dirs.wrf00dir, "wrfinput_d03"), join(ensdir, "wrfinput_d03"))
		server.CopyFile(s.Workdir, join(dirs.wrf00dir, "wrfbdy_d01"), join(ensdir, "wrfbdy_d01"))
	}

	// execute control forecast and all ensemble members
	failed := runForecast(s)

	if <-failed {
		log.Warning("One or more members of the forecast failed to run.")
		return
	}

	log.Info("Post-processing results.")

	log.Info("Simulation completed successfully.")

}

func (s *Simulation) createSimulationDirectories() {
	firstDomain := 1
	if conf.Values.AssimilateOnlyInnerDomain {
		firstDomain = 3
	}
	s.createWrfControlForecastDir(s.Start, s.Duration)
	if conf.Values.AssimilateObservations {
		s.createWrfStepDir(s.Start.Add(-6 * time.Hour))
		s.createWrfStepDir(s.Start.Add(-3 * time.Hour))
		for domain := firstDomain; domain <= 3; domain++ {
			s.createDaDir(s.Start.Add(-6*time.Hour), domain)
			s.createDaDir(s.Start.Add(-3*time.Hour), domain)
			s.createDaDir(s.Start, domain)
		}
	}

	if conf.Values.RunWPS {

		var start time.Time
		var duration time.Duration
		if conf.Values.AssimilateObservations {
			start = s.Start.Add(-6 * time.Hour)
			duration = s.Duration + 6*time.Hour
		} else {
			start = s.Start
			duration = s.Duration
		}

		s.createWpsDir(start, duration)
	}
}

func simDirs(s *Simulation) SimDirs {
	dirs := SimDirs{
		wpsdir:        folders.WPSProcWorkdir(s.Workdir),
		wrf18dir:      folders.WrfControlProcWorkdir(s.Workdir, s.Start.Add(-6*time.Hour)),
		wrf21dir:      folders.WrfControlProcWorkdir(s.Workdir, s.Start.Add(-3*time.Hour)),
		wrf00dir:      folders.WrfControlProcWorkdir(s.Workdir, s.Start),
		wpsOutputsDir: folders.WPSOutputsDir(s.Start),

		da18dir: []string{
			"",
			folders.DAProcWorkdir(s.Workdir, s.Start.Add(-6*time.Hour), 1),
			folders.DAProcWorkdir(s.Workdir, s.Start.Add(-6*time.Hour), 2),
			folders.DAProcWorkdir(s.Workdir, s.Start.Add(-6*time.Hour), 3),
		},
		da21dir: []string{
			"",
			folders.DAProcWorkdir(s.Workdir, s.Start.Add(-3*time.Hour), 1),
			folders.DAProcWorkdir(s.Workdir, s.Start.Add(-3*time.Hour), 2),
			folders.DAProcWorkdir(s.Workdir, s.Start.Add(-3*time.Hour), 3),
		},
		da00dir: []string{
			"",
			folders.DAProcWorkdir(s.Workdir, s.Start, 1),
			folders.DAProcWorkdir(s.Workdir, s.Start, 2),
			folders.DAProcWorkdir(s.Workdir, s.Start, 3),
		},
	}
	return dirs
}

func runForecast(s *Simulation) chan bool {
	failed := make(chan bool, conf.Values.EnsembleMembers)

	go func() {
		var w par.Work[int]
		for ensnum := 0; ensnum <= conf.Values.EnsembleMembers; ensnum++ {
			w.Add(ensnum)
		}
		outfLogPath := filepath.Join(s.Workdir, "output_files.log")

		if err := os.Remove(outfLogPath); err != nil {
			log.Warning("Cannot remove %s: %s", outfLogPath, err)
		}

		w.Do(conf.Values.EnsembleParallelism, func(ensnum int) {
			err := s.RunWrfEnsemble(s.Start, ensnum)
			if err != nil {
				log.Error("Member %d failed: %s", ensnum, err)
				failed <- true
			}
		})
		close(failed)
	}()
	return failed
}

func RunForecastsFromInputs() {
	nodesStr, ok := os.LookupEnv("SLURM_NODELIST")
	if !ok {
		fmt.Fprintln(os.Stderr, "$SLURM_NODELIST not set")
		os.Exit(1)
	}

	nodes, err := mpiman.ParseSlurmNodes(nodesStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot parse $SLURM_NODELIST: %s\n", err)
		os.Exit(1)
	}

	for _, run := range readArgumentsFile() {
		errors.Check(os.Setenv("START_FORECAST", run.start.Format(ShortDtFormat)))
		errors.Check(os.Setenv("DURATION_HOURS", fmt.Sprintf("%.0f", run.duration.Hours())))
		sim := new(run.start, run.duration, nodes)
		sim.run()
	}
}

type run struct {
	start    time.Time
	duration time.Duration
}

func readArgumentsFile() []run {
	var runs []run

	argfilePath := filepath.Join(folders.WPSOutputsRootDir(), "arguments.txt")
	argFile := errors.CheckResult(os.Open(argfilePath))
	defer argFile.Close()
	argReader := bufio.NewReader(argFile)

	// ignore first line, it's the config file name and is not used.
	readline(argReader)

	for {
		line := readline(argReader)
		if len(line) == 0 {
			break
		}
		line = strings.TrimSuffix(line, "\n")
		start := errors.CheckResult(time.Parse("2006010215", line[0:10]))
		duration := time.Hour * time.Duration(errors.CheckResult(strconv.Atoi(line[11:])))
		runs = append(runs, run{start, duration})
	}
	return runs
}

func readline(argReader *bufio.Reader) string {
	line, err := argReader.ReadString('\n')
	if err == io.EOF {
		err = nil
	}
	errors.Check(err)

	return line
}

func RunForecastFromEnv() {
	start := errors.CheckResult(time.Parse(ShortDtFormat, os.Getenv("START_FORECAST")))
	duration := errors.CheckResult(time.ParseDuration(os.Getenv("DURATION_HOURS") + "h"))

	nodesStr, ok := os.LookupEnv("SLURM_NODELIST")
	if !ok {
		fmt.Fprintln(os.Stderr, "$SLURM_NODELIST not set")
		os.Exit(1)
	}

	nodes, err := mpiman.ParseSlurmNodes(nodesStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot parse $SLURM_NODELIST: %s\n", err)
		os.Exit(1)
	}

	sim := new(start, duration, nodes)
	sim.run()
}

func new(start time.Time, duration time.Duration, nodes mpiman.SlurmNodes) Simulation {
	workdir := Workdir(start)

	sim := Simulation{
		Start:    start,
		Duration: duration,
		Workdir:  workdir,
		Nodes:    nodes,
	}
	return sim
}

func Workdir(start time.Time) string {
	workdir := join(folders.WorkDir, start.Format(ShortDtFormat))
	return workdir
}

func (s Simulation) createWpsDir(start time.Time, duration time.Duration) {
	server.RenderTemplate(folders.WPSProcWorkdir(s.Workdir), "wps", start, int(duration.Hours()))
}

func (s Simulation) createWrfControlForecastDir(start time.Time, duration time.Duration) {
	server.RenderTemplate(folders.WrfControlProcWorkdir(s.Workdir, start), "wrf-forecast", start, int(duration.Hours()))
}
func (s Simulation) createWrfEnsembleMemberDir(start time.Time, duration time.Duration, ensnum int) {
	server.RenderTemplate(folders.WrfEnsembleProcWorkdir(s.Workdir, start, ensnum), "wrf-ensmember", start, int(duration.Hours()))
}

func (s Simulation) createWrfStepDir(start time.Time) {
	server.RenderTemplate(folders.WrfControlProcWorkdir(s.Workdir, start), "wrf-step", start, 3)
}

func (s Simulation) createDaDir(start time.Time, domain int) {
	server.RenderTemplate(folders.DAProcWorkdir(s.Workdir, start, domain), fmt.Sprintf("wrfda_%02d", domain), start, 3)
}
