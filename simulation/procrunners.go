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
	"github.com/meteocima/ensemble-runner/wrfprocs"
)

func (s Simulation) RunGeogrid() {
	wpsPath := folders.WPSProcWorkdir(s.Workdir)
	wpsRelDir := errors.CheckResult(filepath.Rel(s.Workdir, wpsPath))

	log.Info("Running geogrid.\t\t\tDIR: $WORKDIR/%s LOGS: %s", wpsRelDir, "geogrid.detail.log geogrid.log.*")
	server.ExecRetry(fmt.Sprintf("mpiexec %s -n %d ./geogrid.exe", conf.Values.MpiOptions, conf.Values.GeogridProcCount), wpsPath, "geogrid.detail.log", "{geogrid.detail.log,geogrid.log.????}")

	logf := errors.CheckResult(os.Open(join(wpsPath, "geogrid.log.0000")))
	defer logf.Close()

	prgs := wrfprocs.ShowGeogridProgress(logf, time.Time{}, time.Time{}.Add(time.Hour))

	var p wrfprocs.Progress
	for p = range prgs {
	}

	if p.Completed {
		if p.Err != nil {
			errors.FailF("geogrid process failed: %w", p.Err)
		} else {
			log.Info("  - Geogrid process completed successfully.")
		}
	}
}

func (s Simulation) RunLinkGrib(startTime time.Time) {
	wpsPath := folders.WPSProcWorkdir(s.Workdir)
	wpsRelDir := errors.CheckResult(filepath.Rel(s.Workdir, wpsPath))

	remoteGfsPath := join(conf.Values.GfsDir, startTime.Format("2006/01/02/1504"))
	log.Info("Running link_grib.\t\t\tDIR: $WORKDIR/%s LOGS: %s", wpsRelDir, "link_grib.detail.log")
	linkCmd := "./link_grib.csh " + remoteGfsPath + "/*.grb"
	server.ExecRetry(linkCmd, wpsPath, "link_grib.detail.log", "link_grib.detail.log")
}

func (s Simulation) RunUngrib() {
	wpsPath := folders.WPSProcWorkdir(s.Workdir)
	wpsRelDir := errors.CheckResult(filepath.Rel(s.Workdir, wpsPath))

	log.Info("Running ungrib.\t\t\t\tDIR: $WORKDIR/%s LOGS: %s", wpsRelDir, "ungrib.detail.log ungrib.log")
	server.ExecRetry("./ungrib.exe", wpsPath, "ungrib.detail.log", "{ungrib.detail.log,ungrib.log}")

	logf := errors.CheckResult(os.Open(join(wpsPath, "ungrib.log")))
	defer logf.Close()

	prgs := wrfprocs.ShowUngribProgress(logf, time.Time{}, time.Time{}.Add(time.Hour))

	var p wrfprocs.Progress
	for p = range prgs {
	}

	if p.Completed {
		if p.Err != nil {
			errors.FailF("ungrib process failed: %w", p.Err)
		} else {
			log.Info("  - Ungrib process completed successfully.")
		}
	}

}

func (s Simulation) RunMetgrid() {
	wpsPath := folders.WPSProcWorkdir(s.Workdir)
	wpsRelDir := errors.CheckResult(filepath.Rel(s.Workdir, wpsPath))

	log.Info("Running metgrid.\t\t\tDIR: $WORKDIR/%s LOGS: %s", wpsRelDir, "metgrid.detail.log metgrid.log.*")
	server.ExecRetry(fmt.Sprintf("mpiexec %s -n %d ./metgrid.exe", conf.Values.MpiOptions, conf.Values.MetgridProcCount), wpsPath, "metgrid.detail.log", "{metgrid.detail.log,metgrid.log.????}")

	logf := errors.CheckResult(os.Open(join(wpsPath, "metgrid.log.0000")))
	defer logf.Close()

	prgs := wrfprocs.ShowMetgridProgress(logf, time.Time{}, time.Time{}.Add(time.Hour))

	var p wrfprocs.Progress
	for p = range prgs {
	}

	if p.Completed {
		if p.Err != nil {
			errors.FailF("metgrid process failed: %w", p.Err)
		} else {
			log.Info("  - Metgrid process completed successfully.")
		}
	}
}

func (s Simulation) RunAvgtsfc() {
	wpsPath := folders.WPSProcWorkdir(s.Workdir)
	wpsRelDir := errors.CheckResult(filepath.Rel(s.Workdir, wpsPath))

	log.Info("Running avg_tsfc.\t\t\tDIR: $WORKDIR/%s LOGS: %s", wpsRelDir, "avg_tsfc.detail.log")
	server.ExecRetry("./avg_tsfc.exe", wpsPath, "avg_tsfc.detail.log", "avg_tsfc.detail.log")
}

func (s Simulation) RunReal(startTime time.Time) {
	wpsPath := folders.WPSProcWorkdir(s.Workdir)
	wpsRelDir := errors.CheckResult(filepath.Rel(s.Workdir, wpsPath))

	log.Info("Running real for %02d:00\t\t\tDIR: $WORKDIR/%s LOGS: %s", startTime.Hour(), wpsRelDir, "real.detail.log,rsl.out.* rsl.error.*")
	server.ExecRetry(fmt.Sprintf("mpiexec %s -n %d ./real.exe", conf.Values.MpiOptions, conf.Values.RealProcCount), wpsPath, "real.detail.log", "{real.detail.log,rsl.out.????,rsl.error.????}")

	logf := errors.CheckResult(os.Open(join(wpsPath, "rsl.out.0000")))
	defer logf.Close()

	prgs := wrfprocs.ShowRealProgress(logf, time.Time{}, time.Time{}.Add(time.Hour))

	var p wrfprocs.Progress
	for p = range prgs {
	}

	if p.Completed {
		if p.Err != nil {
			errors.FailF("real process failed: %w", p.Err)
		} else {
			log.Info("  - Real process completed successfully.")
		}
	}
}

func (s Simulation) RunDa(startTime time.Time, domain int) {

	pathDA := folders.DAProcWorkdir(s.Workdir, startTime, domain)

	daRelDir := errors.CheckResult(filepath.Rel(s.Workdir, pathDA))
	log.Info("Running da_wrfvar for %02d:00 (domain %d)\t\tDIR: $WORKDIR/%s LOGS: %s", startTime.Hour(), domain, daRelDir, "da_wrfvar.detail.log rsl.out.* rsl.error.*")

	server.ExecRetry(fmt.Sprintf("mpirun %s -n %d ./da_wrfvar.exe", conf.Values.MpiOptions, conf.Values.WrfdaProcCount), pathDA, "da_wrfvar.detail.log", "{da_wrfvar.detail.log,rsl.out.????,rsl.error.????}")
	log.Info("  - Da_wrfvar process completed successfully.")
}

func (s Simulation) RunWrf(startTime time.Time, ensnum int) (err error) {
	var path string
	var descr string
	defer errors.OnFailuresSet(&err)
	if ensnum == 0 {
		path = folders.WrfControlProcWorkdir(s.Workdir, startTime)
		descr = "control"
	} else {
		path = folders.WrfEnsembleProcWorkdir(s.Workdir, startTime, ensnum)
		descr = fmt.Sprintf("ensemble n. %d", ensnum)
	}

	wrfRelDir := errors.CheckResult(filepath.Rel(s.Workdir, path))

	log.Info("Running WRF %s for %02d:00\tDIR: $WORKDIR/%s LOGS: %s", descr, startTime.Hour(), wrfRelDir, "wrf.detail.log rsl.out.* rsl.error.*")

	server.ExecRetry(fmt.Sprintf("mpirun %s -n %d ./wrf.exe", conf.Values.MpiOptions, conf.Values.WrfProcCount), path, "wrf.detail.log", "{wrf.detail.log,rsl.out.????,rsl.error.????}")

	logf := errors.CheckResult(os.Open(join(path, "rsl.out.0000")))
	defer logf.Close()

	prgs := wrfprocs.ShowProgress(logf, time.Time{}, time.Time{}.Add(time.Hour))

	var p wrfprocs.Progress
	for p = range prgs {
	}

	if p.Completed {
		if p.Err != nil {
			errors.FailF("WRF %s process failed: %w", descr, p.Err)
		} else {
			log.Info("  - WRF %s process completed successfully.", descr)
		}
	}

	return nil
}
