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
	// define directory for the various steps of the simulation:
	wpsdir := folders.WPSProcWorkdir(s.Workdir)
	wrf18dir := folders.WrfProcWorkdir(s.Workdir, s.Start.Add(-6*time.Hour))
	wrf21dir := folders.WrfProcWorkdir(s.Workdir, s.Start.Add(-3*time.Hour))
	wrf00dir := folders.WrfProcWorkdir(s.Workdir, s.Start)

	// da dirs have one dir for every domain
	da18dir := []string{
		folders.DAProcWorkdir(s.Workdir, s.Start.Add(-6*time.Hour), 1),
		folders.DAProcWorkdir(s.Workdir, s.Start.Add(-6*time.Hour), 2),
		folders.DAProcWorkdir(s.Workdir, s.Start.Add(-6*time.Hour), 3),
	}
	da21dir := []string{
		folders.DAProcWorkdir(s.Workdir, s.Start.Add(-3*time.Hour), 1),
		folders.DAProcWorkdir(s.Workdir, s.Start.Add(-3*time.Hour), 2),
		folders.DAProcWorkdir(s.Workdir, s.Start.Add(-3*time.Hour), 3),
	}
	da00dir := []string{
		folders.DAProcWorkdir(s.Workdir, s.Start, 1),
		folders.DAProcWorkdir(s.Workdir, s.Start, 2),
		folders.DAProcWorkdir(s.Workdir, s.Start, 3),
	}

	// start simulation
	log.Info("Starting simulation from %s for %.0f hours", s.Start.Format(ShortDtFormat), s.Duration.Hours())

	if server.DirExists(s.Workdir) {
		server.Rmdir(s.Workdir)
	}

	defer errors.OnFailuresDo(func(err errors.RunTimeError) {
		panic(err)
	})

	// WPS: create directory wps and run geogrid, ungrib, metgrid
	s.CreateWpsDir(s.Start, s.Duration)
	s.RunGeogrid()
	s.RunLinkGrib(s.Start.Add(-6 * time.Hour))
	s.RunUngrib()
	s.RunAvgtsfc()
	s.RunMetgrid()

	// create directory wrf18 and run real with its namelist
	s.CreateWrfStepDir(s.Start.Add(-6*time.Hour), 3*time.Hour)
	server.CopyFile(join(wrf18dir, "namelist.input"), join(wpsdir, "namelist.input"))
	s.RunReal(s.Start.Add(-6 * time.Hour))

	// assimilate D-6 (first cycle) for all domains
	for domain := 0; domain < 3; domain++ {
		s.CreateDaDir(s.Start.Add(-6*time.Hour), 3, domain+1)
		if domain == 0 {
			server.CopyFile(join(wpsdir, "wrfbdy_d01"), join(da18dir[0], "wrfbdy_d01"))
		}
		// input condition are copied from wps
		server.CopyFile(join(wpsdir, fmt.Sprintf("wrfinput_d%02d", domain+1)), join(da18dir[domain], "fg"))
		s.RunDa(s.Start.Add(-6*time.Hour), domain+1)
	}

	// run WRF from D-6 to D-3. input condition are copied from da dirs of first cycle.
	// boundary condition are copied from wps
	server.CopyFile(join(wpsdir, "wrfbdy_d01"), join(wrf18dir, "wrfbdy_d01"))
	for domain := 0; domain < 3; domain++ {
		server.CopyFile(join(da18dir[domain], "wrfvar_output"), join(wrf18dir, fmt.Sprintf("wrfinput_d%02d", domain+1)))
	}
	s.RunWrf(s.Start.Add(-6 * time.Hour))

	// assimilate D-3 (second cycle) for all domains
	for domain := 0; domain < 3; domain++ {
		s.CreateDaDir(s.Start.Add(-3*time.Hour), 3, domain+1)
		if domain == 0 {
			server.CopyFile(join(wpsdir, "wrfbdy_d01"), join(da21dir[0], "wrfbdy_d01"))
		}
		// input condition are copied from previous cycle wrf
		server.CopyFile(join(wrf18dir, fmt.Sprintf("wrfvar_input_d%02d", domain+1)), join(da21dir[domain], "fg"))
		s.RunDa(s.Start.Add(-3*time.Hour), domain+1)
	}

	// create directory wrf21 and run real with its namelist
	s.CreateWrfStepDir(s.Start.Add(-3*time.Hour), 3*time.Hour)
	server.CopyFile(join(wrf21dir, "namelist.input"), join(folders.WPSProcWorkdir(s.Workdir), "namelist.input"))
	s.RunReal(s.Start.Add(-3 * time.Hour))

	// run WRF from D-3 to D. input conditions are copied from da dirs of second cycle.
	// boundary condition are copied from wps
	server.CopyFile(join(wpsdir, "wrfbdy_d01"), join(wrf21dir, "wrfbdy_d01"))
	for domain := 0; domain < 3; domain++ {
		server.CopyFile(join(da21dir[domain], "wrfvar_output"), join(wrf21dir, fmt.Sprintf("wrfinput_d%02d", domain+1)))
	}
	s.RunWrf(s.Start.Add(-3 * time.Hour))

	// assimilate D (third cycle) for all domains
	for domain := 0; domain < 3; domain++ {
		s.CreateDaDir(s.Start, 3, domain+1)
		if domain == 0 {
			server.CopyFile(join(wpsdir, "wrfbdy_d01"), join(da00dir[0], "wrfbdy_d01"))
		}
		// input condition are copied from previous cycle wrf
		server.CopyFile(join(wrf21dir, fmt.Sprintf("wrfvar_input_d%02d", domain+1)), join(da00dir[domain], "fg"))
		s.RunDa(s.Start, domain+1)
	}

	// create directory wrf00 and run real with its namelist
	s.CreateWrfForecastDir(s.Start, s.Duration)
	server.CopyFile(join(wrf00dir, "namelist.input"), join(wpsdir, "namelist.input"))
	s.RunReal(s.Start)

	// run WRF from D for the duration of the forecast. input conditions are copied from da dirs of third cycle.
	// boundary condition are copied from wps
	server.CopyFile(join(wpsdir, "wrfbdy_d01"), join(wrf00dir, "wrfbdy_d01"))
	for domain := 0; domain < 3; domain++ {
		server.CopyFile(join(da00dir[domain], "wrfvar_output"), join(wrf00dir, fmt.Sprintf("wrfinput_d%02d", domain+1)))
	}
	s.RunWrf(s.Start)
	log.Info("Simulation completed successfully.")
}

func (s Simulation) RunMetgrid() {
	wpsPath := folders.WPSProcWorkdir(s.Workdir)
	wpsRelDir := errors.CheckResult(filepath.Rel(folders.Rootdir, wpsPath))

	log.Info("Running metgrid.\t\t\tDIR: %s LOGS: %s", wpsRelDir, "metgrid.detail.log metgrid.log.*")
	server.ExecRetry(fmt.Sprintf("mpiexec %s -n %d ./metgrid.exe", conf.Values.MpiOptions, conf.Values.MetgridProcCount), wpsPath, "metgrid.detail.log", "{metgrid.detail.log,metgrid.log.????}")
}

func (s Simulation) RunAvgtsfc() {
	wpsPath := folders.WPSProcWorkdir(s.Workdir)
	wpsRelDir := errors.CheckResult(filepath.Rel(folders.Rootdir, wpsPath))

	log.Info("Running avg_tsfc.\t\t\tDIR: %s LOGS: %s", wpsRelDir, "avg_tsfc.detail.log")
	server.ExecRetry("./avg_tsfc.exe", wpsPath, "avg_tsfc.detail.log", "avg_tsfc.detail.log")
}

func (s Simulation) RunUngrib() {
	wpsPath := folders.WPSProcWorkdir(s.Workdir)
	wpsRelDir := errors.CheckResult(filepath.Rel(folders.Rootdir, wpsPath))

	log.Info("Running ungrib.\t\t\t\tDIR: %s LOGS: %s", wpsRelDir, "ungrib.detail.log ungrib.log")
	server.ExecRetry("./ungrib.exe", wpsPath, "ungrib.detail.log", "{ungrib.detail.log,ungrib.log}")
}

func (s Simulation) RunLinkGrib(startTime time.Time) {
	wpsPath := folders.WPSProcWorkdir(s.Workdir)
	wpsRelDir := errors.CheckResult(filepath.Rel(folders.Rootdir, wpsPath))

	remoteGfsPath := join(conf.Values.GfsDir, startTime.Format("2006/01/02/1504"))
	log.Info("Running link_grib.\t\t\tDIR: %s LOGS: %s", wpsRelDir, "link_grib.detail.log")
	linkCmd := "./link_grib.csh " + remoteGfsPath + "/*.grb"
	server.ExecRetry(linkCmd, wpsPath, "link_grib.detail.log", "link_grib.detail.log")
}

func (s Simulation) RunGeogrid() {
	wpsPath := folders.WPSProcWorkdir(s.Workdir)
	wpsRelDir := errors.CheckResult(filepath.Rel(folders.Rootdir, wpsPath))

	log.Info("Running geogrid.\t\t\tDIR: %s LOGS: %s", wpsRelDir, "geogrid.detail.log geogrid.log.*")
	server.ExecRetry(fmt.Sprintf("mpiexec %s -n %d ./geogrid.exe", conf.Values.MpiOptions, conf.Values.GeogridProcCount), wpsPath, "geogrid.detail.log", "{geogrid.detail.log,geogrid.log.????}")
}

func (s Simulation) RunReal(startTime time.Time) {
	wpsPath := folders.WPSProcWorkdir(s.Workdir)
	wpsRelDir := errors.CheckResult(filepath.Rel(folders.Rootdir, wpsPath))

	log.Info("Running real for %02d:00\t\t\tDIR: %s LOGS: %s", startTime.Hour(), wpsRelDir, "real.detail.log,rsl.out.* rsl.error.*")
	server.ExecRetry(fmt.Sprintf("mpiexec %s -n %d ./real.exe", conf.Values.MpiOptions, conf.Values.RealProcCount), wpsPath, "real.detail.log", "{real.detail.log,rsl.out.????,rsl.error.????}")
}

func (s Simulation) RunDa(startTime time.Time, domain int) {

	pathDA := folders.DAProcWorkdir(s.Workdir, startTime, domain)

	daRelDir := errors.CheckResult(filepath.Rel(folders.Rootdir, pathDA))
	log.Info("Running da_wrfvar for %02d:00 (domain %d)\t\tDIR: %s LOGS: %s", startTime.Hour(), domain, daRelDir, "da_wrfvar.detail.log rsl.out.* rsl.error.*")

	server.ExecRetry(fmt.Sprintf("mpirun %s -n %d ./da_wrfvar.exe", conf.Values.MpiOptions, conf.Values.WrfdaProcCount), pathDA, "da_wrfvar.detail.log", "{da_wrfvar.detail.log,rsl.out.????,rsl.error.????}")
}

func (s Simulation) RunWrf(startTime time.Time) string {
	path := folders.WrfProcWorkdir(s.Workdir, startTime)

	wrfRelDir := errors.CheckResult(filepath.Rel(folders.Rootdir, path))

	log.Info("Running wrf for %02d:00\t\t\tDIR: %s LOGS: %s", startTime.Hour(), wrfRelDir, "wrf.detail.log rsl.out.* rsl.error.*")

	server.ExecRetry(fmt.Sprintf("mpirun %s -n %d ./wrf.exe", conf.Values.MpiOptions, conf.Values.WrfProcCount), path, "wrf.detail.log", "{wrf.detail.log,rsl.out.????,rsl.error.????}")

	return path
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

func (s Simulation) CreateWrfStepDir(start time.Time, duration time.Duration) {
	server.RenderTemplate(folders.WrfProcWorkdir(s.Workdir, start), "wrf-step", start, 3)
}

func (s Simulation) CreateDaDir(start time.Time, duration time.Duration, domain int) {
	server.RenderTemplate(folders.DAProcWorkdir(s.Workdir, start, domain), "wrfda_03", start, 3)
}
