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

func (s *Simulation) Run() {
	// define directories vars for the workdir of various steps of the simulation:
	wpsdir := folders.WPSProcWorkdir(s.Workdir)
	wrf18dir := folders.WrfControlProcWorkdir(s.Workdir, s.Start.Add(-6*time.Hour))
	wrf21dir := folders.WrfControlProcWorkdir(s.Workdir, s.Start.Add(-3*time.Hour))
	wrf00dir := folders.WrfControlProcWorkdir(s.Workdir, s.Start)
	wpsOutputsDir := folders.WPSOutputsDir(s.Start)

	// da dirs have one dir for every domain
	da18dir := []string{
		"",
		folders.DAProcWorkdir(s.Workdir, s.Start.Add(-6*time.Hour), 1),
		folders.DAProcWorkdir(s.Workdir, s.Start.Add(-6*time.Hour), 2),
		folders.DAProcWorkdir(s.Workdir, s.Start.Add(-6*time.Hour), 3),
	}
	da21dir := []string{
		"",
		folders.DAProcWorkdir(s.Workdir, s.Start.Add(-3*time.Hour), 1),
		folders.DAProcWorkdir(s.Workdir, s.Start.Add(-3*time.Hour), 2),
		folders.DAProcWorkdir(s.Workdir, s.Start.Add(-3*time.Hour), 3),
	}
	da00dir := []string{
		"",
		folders.DAProcWorkdir(s.Workdir, s.Start, 1),
		folders.DAProcWorkdir(s.Workdir, s.Start, 2),
		folders.DAProcWorkdir(s.Workdir, s.Start, 3),
	}

	// start simulation
	log.Info("Starting simulation from %s for %.0f hours", s.Start.Format(ShortDtFormat), s.Duration.Hours())
	log.Info("  -- $WORKDIR=%s", s.Workdir)

	if server.DirExists(s.Workdir) {
		server.Rmdir(s.Workdir)
	}

	defer errors.OnFailuresDo(func(err errors.RunTimeError) {
		panic(err)
	})

	firstDomain := 1
	if conf.Values.AssimilateOnlyInnerDomain {
		firstDomain = 3
	}

	// create all directories for the various wrf cycles.
	s.CreateWrfControlForecastDir(s.Start, s.Duration)
	if conf.Values.AssimilateObservations {
		s.CreateWrfStepDir(s.Start.Add(-6 * time.Hour))
		s.CreateWrfStepDir(s.Start.Add(-3 * time.Hour))
		for domain := firstDomain; domain <= 3; domain++ {
			s.CreateDaDir(s.Start.Add(-6*time.Hour), domain)
			s.CreateDaDir(s.Start.Add(-3*time.Hour), domain)
			s.CreateDaDir(s.Start, domain)
		}
	}

	if conf.Values.EnsembleMembers > 0 {
		rnd := rand.NewSource(0xfeedbabebadcafe)

		for ensnum := 1; ensnum <= conf.Values.EnsembleMembers; ensnum++ {
			seed := rnd.Int63()%100 + 1
			os.Setenv("ENSEMBLE_SEED", fmt.Sprintf("%02d", seed))
			log.Debug("Using seed %02d for member n.%d.", seed, ensnum)

			s.CreateWrfEnsembleMemberDir(s.Start, s.Duration, ensnum)
		}
	}

	if conf.Values.RunWPS {
		// WPS: create directory wps and run geogrid, ungrib, metgrid
		var start time.Time
		var duration time.Duration
		if conf.Values.AssimilateObservations {
			start = s.Start.Add(-6 * time.Hour)
			duration = s.Duration + 6*time.Hour
		} else {
			start = s.Start
			duration = s.Duration
		}

		s.CreateWpsDir(start, duration)
		s.RunGeogrid()

		s.RunLinkGrib(start)
		s.RunUngrib()
		if duration > 24*time.Hour {
			s.RunAvgtsfc()
		}
		s.RunMetgrid()

		server.MkdirAll(wpsOutputsDir, 0775)

		if conf.Values.AssimilateObservations {
			// run real for every cycle.
			server.CopyFile(s.Workdir, join(wrf18dir, "namelist.input"), join(wpsdir, "namelist.input"))
			s.RunReal(s.Start.Add(-6 * time.Hour))
			server.CopyFile(s.Workdir, join(wpsdir, "wrfbdy_d01"), join(wpsOutputsDir, "wrfbdy_d01_da01"))
			server.CopyFile(s.Workdir, join(wpsdir, "wrfinput_d01"), join(wpsOutputsDir, "wrfinput_d01"))
			server.CopyFile(s.Workdir, join(wpsdir, "wrfinput_d02"), join(wpsOutputsDir, "wrfinput_d02"))
			server.CopyFile(s.Workdir, join(wpsdir, "wrfinput_d03"), join(wpsOutputsDir, "wrfinput_d03"))

			server.CopyFile(s.Workdir, join(wrf21dir, "namelist.input"), join(wpsdir, "namelist.input"))
			s.RunReal(s.Start.Add(-3 * time.Hour))
			server.CopyFile(s.Workdir, join(wpsdir, "wrfbdy_d01"), join(wpsOutputsDir, "wrfbdy_d01_da02"))

			server.CopyFile(s.Workdir, join(wrf00dir, "namelist.input"), join(wpsdir, "namelist.input"))
			s.RunReal(s.Start)
			server.CopyFile(s.Workdir, join(wpsdir, "wrfbdy_d01"), join(wpsOutputsDir, "wrfbdy_d01_da03"))
		} else {
			server.CopyFile(s.Workdir, join(wrf00dir, "namelist.input"), join(wpsdir, "namelist.input"))
			s.RunReal(s.Start)
			server.CopyFile(s.Workdir, join(wpsdir, "wrfbdy_d01"), join(wpsOutputsDir, "wrfbdy_d01"))
			server.CopyFile(s.Workdir, join(wpsdir, "wrfinput_d01"), join(wpsOutputsDir, "wrfinput_d01"))
			server.CopyFile(s.Workdir, join(wpsdir, "wrfinput_d02"), join(wpsOutputsDir, "wrfinput_d02"))
			server.CopyFile(s.Workdir, join(wpsdir, "wrfinput_d03"), join(wpsOutputsDir, "wrfinput_d03"))
		}

	}

	if conf.Values.AssimilateObservations {
		if conf.Values.AssimilateFirstCycle {
			// assimilate D-6 (first cycle) for all domains
			if !conf.Values.AssimilateOnlyInnerDomain {
				server.CopyFile(s.Workdir, join(wpsOutputsDir, "wrfbdy_d01_da01"), join(da18dir[1], "wrfbdy_d01"))
			}
			for domain := firstDomain; domain <= 3; domain++ {
				// input condition are copied from wps
				server.CopyFile(s.Workdir, join(wpsOutputsDir, fmt.Sprintf("wrfinput_d%02d", domain)), join(da18dir[domain], "fg"))
				s.RunDa(s.Start.Add(-6*time.Hour), domain)
			}

			// to run WRF from D-6 to D-3, input and boundary conditions are copied from da dirs of first cycle.
			if !conf.Values.AssimilateOnlyInnerDomain {
				server.CopyFile(s.Workdir, join(da18dir[1], "wrfbdy_d01"), join(wrf18dir, "wrfbdy_d01"))
			} else {
				server.CopyFile(s.Workdir, join(wpsOutputsDir, "wrfbdy_d01_da01"), join(wrf18dir, "wrfbdy_d01"))
			}
			for domain := firstDomain; domain <= 3; domain++ {
				server.CopyFile(s.Workdir, join(da18dir[domain], "wrfvar_output"), join(wrf18dir, fmt.Sprintf("wrfinput_d%02d", domain)))
			}
			if conf.Values.AssimilateOnlyInnerDomain {
				for domain := 1; domain <= 2; domain++ {
					server.CopyFile(s.Workdir, join(wpsOutputsDir, fmt.Sprintf("wrfinput_d%02d", domain)), join(wrf18dir, fmt.Sprintf("wrfinput_d%02d", domain)))
				}
			}

		} else {
			// to run WRF from D-6 to D-3, input and boundary conditions are copied from wps.
			server.CopyFile(s.Workdir, join(wpsOutputsDir, "wrfbdy_d01_da01"), join(wrf18dir, "wrfbdy_d01"))
			for domain := 1; domain <= 3; domain++ {
				server.CopyFile(s.Workdir, join(wpsOutputsDir, fmt.Sprintf("wrfinput_d%02d", domain)), join(wrf18dir, fmt.Sprintf("wrfinput_d%02d", domain)))
			}
		}
	} else {
		server.CopyFile(s.Workdir, join(wpsOutputsDir, "wrfbdy_d01"), join(wrf00dir, "wrfbdy_d01"))
		for domain := 1; domain <= 3; domain++ {
			server.CopyFile(s.Workdir, join(wpsOutputsDir, fmt.Sprintf("wrfinput_d%02d", domain)), join(wrf00dir, fmt.Sprintf("wrfinput_d%02d", domain)))
		}
	}

	if conf.Values.AssimilateObservations {
		// run WRF from D-6 to D-3.
		s.RunWrfStep(s.Start.Add(-6 * time.Hour))

		// assimilate D-3 (second cycle) for all domains
		if !conf.Values.AssimilateOnlyInnerDomain {
			server.CopyFile(s.Workdir, join(wpsOutputsDir, "wrfbdy_d01_da02"), join(da21dir[1], "wrfbdy_d01"))
		}
		for domain := firstDomain; domain <= 3; domain++ {
			// input condition are copied from previous cycle wrf
			server.CopyFile(s.Workdir, join(wrf18dir, fmt.Sprintf("wrfvar_input_d%02d", domain)), join(da21dir[domain], "fg"))
			s.RunDa(s.Start.Add(-3*time.Hour), domain)
		}

		// run WRF from D-3 to D. input and boundary conditions are copied from da dirs of second cycle.
		// boundary condition are copied from wps
		if conf.Values.AssimilateOnlyInnerDomain {
			server.CopyFile(s.Workdir, join(wpsOutputsDir, "wrfbdy_d01_da02"), join(wrf21dir, "wrfbdy_d01"))
		} else {
			server.CopyFile(s.Workdir, join(da21dir[1], "wrfbdy_d01"), join(wrf21dir, "wrfbdy_d01"))
		}

		for domain := firstDomain; domain <= 3; domain++ {
			server.CopyFile(s.Workdir, join(da21dir[domain], "wrfvar_output"), join(wrf21dir, fmt.Sprintf("wrfinput_d%02d", domain)))
		}
		if conf.Values.AssimilateOnlyInnerDomain {
			for domain := 1; domain <= 2; domain++ {
				server.CopyFile(s.Workdir, join(wrf18dir, fmt.Sprintf("wrfvar_input_d%02d", domain)), join(wrf21dir, fmt.Sprintf("wrfinput_d%02d", domain)))
			}
		}

		s.RunWrfStep(s.Start.Add(-3 * time.Hour))

		// assimilate D (third cycle) for all domains
		if !conf.Values.AssimilateOnlyInnerDomain {
			server.CopyFile(s.Workdir, join(wpsOutputsDir, "wrfbdy_d01_da03"), join(da00dir[1], "wrfbdy_d01"))
		}
		for domain := firstDomain; domain <= 3; domain++ {
			// input condition are copied from previous cycle wrf
			server.CopyFile(s.Workdir, join(wrf21dir, fmt.Sprintf("wrfvar_input_d%02d", domain)), join(da00dir[domain], "fg"))
			s.RunDa(s.Start, domain)
		}

		if conf.Values.AssimilateOnlyInnerDomain {
			server.CopyFile(s.Workdir, join(wpsOutputsDir, "wrfbdy_d01_da03"), join(wrf00dir, "wrfbdy_d01"))
		} else {
			server.CopyFile(s.Workdir, join(da00dir[1], "wrfbdy_d01"), join(wrf00dir, "wrfbdy_d01"))
		}
		// run WRF from D for the duration of the forecast. input and boundary conditions are copied from da dirs of third cycle.
		for domain := firstDomain; domain <= 3; domain++ {
			server.CopyFile(s.Workdir, join(da00dir[domain], "wrfvar_output"), join(wrf00dir, fmt.Sprintf("wrfinput_d%02d", domain)))
		}
		if conf.Values.AssimilateOnlyInnerDomain {
			for domain := 1; domain <= 2; domain++ {
				server.CopyFile(s.Workdir, join(wrf21dir, fmt.Sprintf("wrfvar_input_d%02d", domain)), join(wrf00dir, fmt.Sprintf("wrfinput_d%02d", domain)))
			}
		}
	}

	// if an ensemble is procduced, copy wrfinput and wrfbdy from control forecast to all ensemble members
	for ensnum := 1; ensnum <= conf.Values.EnsembleMembers; ensnum++ {
		ensdir := folders.WrfEnsembleProcWorkdir(s.Workdir, s.Start, ensnum)
		server.CopyFile(s.Workdir, join(wrf00dir, "wrfinput_d01"), join(ensdir, "wrfinput_d01"))
		server.CopyFile(s.Workdir, join(wrf00dir, "wrfinput_d02"), join(ensdir, "wrfinput_d02"))
		server.CopyFile(s.Workdir, join(wrf00dir, "wrfinput_d03"), join(ensdir, "wrfinput_d03"))
		server.CopyFile(s.Workdir, join(wrf00dir, "wrfbdy_d01"), join(ensdir, "wrfbdy_d01"))
	}

	// execute control forecast and all ensemble members
	failed := RunForecast(s)

	if <-failed {
		log.Warning("One or more members of the forecast failed to run.")
		return
	}

	log.Info("Post-processing results.")
	server.ExecRetry(fmt.Sprintf("ROOTDIR='%s' delivery.sh > delivery.log 2>&1", folders.Rootdir), s.Workdir, "delivery.log", "delivery.log")

	log.Info("Simulation completed successfully.")

}

func RunForecast(s *Simulation) chan bool {
	failed := make(chan bool, conf.Values.EnsembleMembers)

	go func() {
		var w par.Work[int]
		for ensnum := 0; ensnum <= conf.Values.EnsembleMembers; ensnum++ {
			w.Add(ensnum)
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
		sim := New(run.start, run.duration, nodes)
		sim.Run()
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

	sim := New(start, duration, nodes)
	sim.Run()
}

func New(start time.Time, duration time.Duration, nodes mpiman.SlurmNodes) Simulation {
	workdir := join(folders.WorkDir, start.Format(ShortDtFormat))

	sim := Simulation{
		Start:    start,
		Duration: duration,
		Workdir:  workdir,
		Nodes:    nodes,
	}
	return sim
}

func (s Simulation) CreateWpsDir(start time.Time, duration time.Duration) {
	server.RenderTemplate(folders.WPSProcWorkdir(s.Workdir), "wps", start, int(duration.Hours()))
}

func (s Simulation) CreateWrfControlForecastDir(start time.Time, duration time.Duration) {
	server.RenderTemplate(folders.WrfControlProcWorkdir(s.Workdir, start), "wrf-forecast", start, int(duration.Hours()))
}
func (s Simulation) CreateWrfEnsembleMemberDir(start time.Time, duration time.Duration, ensnum int) {
	server.RenderTemplate(folders.WrfEnsembleProcWorkdir(s.Workdir, start, ensnum), "wrf-ensmember", start, int(duration.Hours()))
}

func (s Simulation) CreateWrfStepDir(start time.Time) {
	server.RenderTemplate(folders.WrfControlProcWorkdir(s.Workdir, start), "wrf-step", start, 3)
}

func (s Simulation) CreateDaDir(start time.Time, domain int) {
	server.RenderTemplate(folders.DAProcWorkdir(s.Workdir, start, domain), fmt.Sprintf("wrfda_%02d", domain), start, 3)
}
