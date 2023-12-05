package simulation

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/meteocima/ensemble-runner/conf"
	"github.com/meteocima/ensemble-runner/errors"
	"github.com/meteocima/ensemble-runner/folders"
	"github.com/meteocima/ensemble-runner/log"
	"github.com/meteocima/ensemble-runner/server"
)

type Simulation struct {
	Start    time.Time
	Duration time.Duration
	Workdir  string
}

var ShortDtFormat = "2006-01-02-15"

var join = filepath.Join

func (s *Simulation) Run() {
	// define directories vars for the workdir of various steps of the simulation:
	wpsdir := folders.WPSProcWorkdir(s.Workdir)
	wrf18dir := folders.WrfProcWorkdir(s.Workdir, s.Start.Add(-6*time.Hour))
	wrf21dir := folders.WrfProcWorkdir(s.Workdir, s.Start.Add(-3*time.Hour))
	wrf00dir := folders.WrfProcWorkdir(s.Workdir, s.Start)
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
	s.CreateWrfStepDir(s.Start.Add(-6 * time.Hour))
	s.CreateWrfStepDir(s.Start.Add(-3 * time.Hour))
	s.CreateWrfForecastDir(s.Start, s.Duration)
	for domain := firstDomain; domain <= 3; domain++ {
		s.CreateDaDir(s.Start.Add(-6*time.Hour), domain)
		s.CreateDaDir(s.Start.Add(-3*time.Hour), domain)
		s.CreateDaDir(s.Start, domain)
	}

	if conf.Values.RunWPS {
		// WPS: create directory wps and run geogrid, ungrib, metgrid
		s.CreateWpsDir(s.Start, s.Duration)
		s.RunGeogrid()
		s.RunLinkGrib(s.Start.Add(-6 * time.Hour))
		s.RunUngrib()
		if s.Duration+6*time.Hour > 24*time.Hour {
			s.RunAvgtsfc()
		}
		s.RunMetgrid()

		server.MkdirAll(wpsOutputsDir, 0775)

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
	}

	if conf.Values.AssimilateFirstCycle {
		// assimilate D-6 (first cycle) for all domains
		server.CopyFile(s.Workdir, join(wpsOutputsDir, "wrfbdy_d01_da01"), join(da18dir[1], "wrfbdy_d01"))
		for domain := firstDomain; domain <= 3; domain++ {
			// input condition are copied from wps
			server.CopyFile(s.Workdir, join(wpsOutputsDir, fmt.Sprintf("wrfinput_d%02d", domain)), join(da18dir[domain], "fg"))
			s.RunDa(s.Start.Add(-6*time.Hour), domain)
		}

		// to run WRF from D-6 to D-3, input and boundary conditions are copied from da dirs of first cycle.
		if !conf.Values.AssimilateOnlyInnerDomain {
			server.CopyFile(s.Workdir, join(da18dir[1], "wrfbdy_d01"), join(wrf18dir, "wrfbdy_d01"))
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

	// run WRF from D-6 to D-3.
	s.RunWrf(s.Start.Add(-6 * time.Hour))

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

	s.RunWrf(s.Start.Add(-3 * time.Hour))

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
	s.RunWrf(s.Start)
	log.Info("Simulation completed successfully.")
}

func New() Simulation {
	start := errors.CheckResult(time.Parse(ShortDtFormat, os.Getenv("START_FORECAST")))
	duration := errors.CheckResult(time.ParseDuration(os.Getenv("DURATION_HOURS") + "h"))
	workdir := join(folders.WorkDir, os.Getenv("START_FORECAST"))

	sim := Simulation{
		Start:    start,
		Duration: duration,
		Workdir:  workdir,
	}

	return sim
}

func (s Simulation) CreateWpsDir(start time.Time, duration time.Duration) {
	server.RenderTemplate(folders.WPSProcWorkdir(s.Workdir), "wps", start.Add(-6*time.Hour), 6+int(duration.Hours()))
}

func (s Simulation) CreateWrfForecastDir(start time.Time, duration time.Duration) {
	server.RenderTemplate(folders.WrfProcWorkdir(s.Workdir, start), "wrf-forecast", start, int(duration.Hours()))
}

func (s Simulation) CreateWrfStepDir(start time.Time) {
	server.RenderTemplate(folders.WrfProcWorkdir(s.Workdir, start), "wrf-step", start, 3)
}

func (s Simulation) CreateDaDir(start time.Time, domain int) {
	server.RenderTemplate(folders.DAProcWorkdir(s.Workdir, start, domain), fmt.Sprintf("wrfda_%02d", domain), start, 3)
}
