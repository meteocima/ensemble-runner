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

func New() Simulation {
	start := errors.CheckResult(time.Parse(ShortDtFormat, os.Getenv("START_FORECAST")))
	duration := errors.CheckResult(time.ParseDuration(os.Getenv("DURATION_HOURS") + "h"))
	workdir := filepath.Join(folders.WorkDir, os.Getenv("START_FORECAST"))

	sim := Simulation{
		Start:    start,
		Duration: duration,
		Workdir:  workdir,
	}

	return sim
}

func (s Simulation) CreateWPSDir(start time.Time, duration time.Duration) {
	server.RenderTemplate(folders.WPSProcWorkdir(s.Workdir), "wps", start.Add(-6*time.Hour), 6+int(duration.Hours()))
}

func (s Simulation) CreateWRFForecastDir(start time.Time, duration time.Duration) {
	server.RenderTemplate(folders.WrfProcWorkdir(s.Workdir, start), "wrf-forecast", start, int(duration.Hours()))
}

func (s Simulation) CreateWRFStepDir(start time.Time, duration time.Duration) {
	server.RenderTemplate(folders.WrfProcWorkdir(s.Workdir, start), "wrf-step", start, 3)
}

func (s Simulation) CreateDADir(start time.Time, duration time.Duration) {
	server.RenderTemplate(folders.DAProcWorkdir(s.Workdir, start, 3), "wrfda_03", start, 3)
}

func (s *Simulation) Run() {
	log.Info("Starting simulation from %s for %.0f hours", s.Start.Format(ShortDtFormat), s.Duration.Hours())

	if server.DirExists(s.Workdir) {
		server.Rmdir(s.Workdir)
	}

	defer errors.OnFailuresDo(func(err errors.RunTimeError) {
		panic(err)
	})

	s.CreateWPSDir(s.Start, s.Duration)

	// prepare folder and run geogrid, ungrib, metgrid
	s.RunWPS(s.Start.Add(-6*time.Hour), int(s.Duration.Hours()))

	// run WRF from D-6 to D-3
	s.CreateWRFStepDir(s.Start.Add(-6*time.Hour), 3*time.Hour)
	server.CopyFile(
		filepath.Join(folders.WrfProcWorkdir(s.Workdir, s.Start.Add(-6*time.Hour)), "namelist.input"),
		filepath.Join(folders.WPSProcWorkdir(s.Workdir), "namelist.input"),
	)
	s.RunREAL(s.Start.Add(-6*time.Hour), 3)

	for _, file := range []string{"wrfbdy_d01", "wrfinput_d01", "wrfinput_d02", "wrfinput_d03"} {
		server.CopyFile(
			filepath.Join(folders.WPSProcWorkdir(s.Workdir), file),
			filepath.Join(folders.WrfProcWorkdir(s.Workdir, s.Start.Add(-6*time.Hour)), file),
		)
	}

	s.RunWRF(s.Start.Add(-6*time.Hour), 3)

	// assimilate D-3
	s.CreateDADir(s.Start.Add(-3*time.Hour), 3)
	s.RunDA(s.Start.Add(-3*time.Hour), 3)

	// run WRF from D-3 to D
	s.CreateWRFStepDir(s.Start.Add(-3*time.Hour), 3*time.Hour)
	server.CopyFile(
		filepath.Join(folders.WrfProcWorkdir(s.Workdir, s.Start.Add(-3*time.Hour)), "namelist.input"),
		filepath.Join(folders.WPSProcWorkdir(s.Workdir), "namelist.input"),
	)
	s.RunREAL(s.Start.Add(-3*time.Hour), 3)

	for _, file := range []string{"wrfbdy_d01", "wrfinput_d01", "wrfinput_d02"} {
		server.CopyFile(
			filepath.Join(folders.WPSProcWorkdir(s.Workdir), file),
			filepath.Join(folders.WrfProcWorkdir(s.Workdir, s.Start.Add(-3*time.Hour)), file),
		)
	}

	server.CopyFile(
		filepath.Join(folders.DAProcWorkdir(s.Workdir, s.Start.Add(-3*time.Hour), 3), "wrfvar_output"),
		filepath.Join(folders.WrfProcWorkdir(s.Workdir, s.Start.Add(-3*time.Hour)), "wrfinput_d03"),
	)
	s.RunWRF(s.Start.Add(-3*time.Hour), 3)

	// assimilate D
	s.CreateDADir(s.Start, 3)
	s.RunDA(s.Start, 3)

	// run WRF from D for the duration of the forecast
	s.CreateWRFForecastDir(s.Start, s.Duration)

	// run REAL
	server.CopyFile(
		filepath.Join(folders.WrfProcWorkdir(s.Workdir, s.Start), "namelist.input"),
		filepath.Join(folders.WPSProcWorkdir(s.Workdir), "namelist.input"),
	)

	s.RunREAL(s.Start, int(s.Duration.Hours()))

	for _, file := range []string{"wrfbdy_d01", "wrfinput_d01", "wrfinput_d02"} {
		server.CopyFile(
			filepath.Join(folders.WPSProcWorkdir(s.Workdir), file),
			filepath.Join(folders.WrfProcWorkdir(s.Workdir, s.Start), file),
		)
	}

	server.CopyFile(
		filepath.Join(folders.DAProcWorkdir(s.Workdir, s.Start, 3), "wrfvar_output"),
		filepath.Join(folders.WrfProcWorkdir(s.Workdir, s.Start), "wrfinput_d03"),
	)
	s.RunWRF(s.Start, int(s.Duration.Hours()))
	log.Info("Simulation completed successfully.")
}

func (s Simulation) RunWPS(startTime time.Time, duration int) string {

	remoteGfsPath := startTime.Format("/data/unsafe/gfs/2006/01/02/1504/")

	path := folders.WPSProcWorkdir(s.Workdir)
	wpsRelDir := errors.CheckResult(filepath.Rel(folders.Rootdir, path))
	log.Info("Running geogrid.\t\t\tDIR: %s LOGS: %s", wpsRelDir, "geogrid.detail.log geogrid.log.*")
	server.ExecRetry(fmt.Sprintf("mpiexec %s -n %d ./geogrid.exe", conf.Values.MpiOptions, conf.Values.GeogridProcCount), path, "geogrid.detail.log", "{geogrid.detail.log,geogrid.log.????}")

	log.Info("Running link_grib.\t\t\tDIR: %s LOGS: %s", wpsRelDir, "link_grib.detail.log")
	linkCmd := "./link_grib.csh " + remoteGfsPath + "/*.grb"
	server.ExecRetry(linkCmd, path, "link_grib.detail.log", "link_grib.detail.log")

	log.Info("Running ungrib.\t\t\t\tDIR: %s LOGS: %s", wpsRelDir, "ungrib.detail.log ungrib.log")
	server.ExecRetry("./ungrib.exe", path, "ungrib.detail.log", "{ungrib.detail.log,ungrib.log}")

	log.Info("Running avg_tsfc.\t\t\tDIR: %s LOGS: %s", wpsRelDir, "avg_tsfc.detail.log")
	server.ExecRetry("./avg_tsfc.exe", path, "avg_tsfc.detail.log", "avg_tsfc.detail.log")

	log.Info("Running metgrid.\t\t\tDIR: %s LOGS: %s", wpsRelDir, "metgrid.detail.log metgrid.log.*")
	server.ExecRetry(fmt.Sprintf("mpiexec %s -n %d ./metgrid.exe", conf.Values.MpiOptions, conf.Values.MetgridProcCount), path, "metgrid.detail.log", "{metgrid.detail.log,metgrid.log.????}")

	return path
}

func (s Simulation) RunREAL(startTime time.Time, duration int) {
	wpsPath := folders.WPSProcWorkdir(s.Workdir)
	wpsRelDir := errors.CheckResult(filepath.Rel(folders.Rootdir, wpsPath))

	log.Info("Running real for %02d:00\t\t\tDIR: %s LOGS: %s", startTime.Hour(), wpsRelDir, "real.detail.log,rsl.out.* rsl.error.*")
	server.ExecRetry(fmt.Sprintf("mpiexec %s -n %d ./real.exe", conf.Values.MpiOptions, conf.Values.RealProcCount), wpsPath, "real.detail.log", "{real.detail.log,rsl.out.????,rsl.error.????}")
}

func (s Simulation) RunDA(startTime time.Time, duration int) {

	pathDA := folders.DAProcWorkdir(s.Workdir, startTime, 3)
	inputPath := folders.WrfProcWorkdir(s.Workdir, startTime.Add(-3*time.Hour))
	src := filepath.Join(inputPath, "wrfvar_input_d03")
	dest := filepath.Join(pathDA, "fg")
	server.CopyFile(src, dest)
	daRelDir := errors.CheckResult(filepath.Rel(folders.Rootdir, pathDA))
	log.Info("Running da_wrfvar for %02d:00\t\tDIR: %s LOGS: %s", startTime.Hour(), daRelDir, "da_wrfvar.detail.log rsl.out.* rsl.error.*")

	server.ExecRetry(fmt.Sprintf("mpirun %s -n %d ./da_wrfvar.exe", conf.Values.MpiOptions, conf.Values.WrfdaProcCount), pathDA, "da_wrfvar.detail.log", "{da_wrfvar.detail.log,rsl.out.????,rsl.error.????}")
}

func (s Simulation) RunWRF(startTime time.Time, duration int) string {
	path := folders.WrfProcWorkdir(s.Workdir, startTime)

	wrfRelDir := errors.CheckResult(filepath.Rel(folders.Rootdir, path))

	log.Info("Running wrf for %02d:00\t\t\tDIR: %s LOGS: %s", startTime.Hour(), wrfRelDir, "wrf.detail.log rsl.out.* rsl.error.*")

	server.ExecRetry(fmt.Sprintf("mpirun %s -n %d ./wrf.exe", conf.Values.MpiOptions, conf.Values.WrfProcCount), path, "wrf.detail.log", "{wrf.detail.log,rsl.out.????,rsl.error.????}")

	return path
}
