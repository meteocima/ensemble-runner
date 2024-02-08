package simulation

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/meteocima/ensemble-runner/conf"
	"github.com/meteocima/ensemble-runner/errors"
	"github.com/meteocima/ensemble-runner/folders"
	"github.com/meteocima/ensemble-runner/log"
	"github.com/meteocima/ensemble-runner/mpiman"
	"github.com/meteocima/ensemble-runner/server"
	"github.com/meteocima/ensemble-runner/wrfprocs"
	"github.com/parro-it/tailor"
)

func (s Simulation) RunGeogrid() {
	wpsPath := folders.WPSProcWorkdir(s.Workdir)
	wpsRelDir := errors.CheckResult(filepath.Rel(s.Workdir, wpsPath))

	log.Info("Running geogrid.\t\t\tDIR: $WORKDIR/%s LOGS: %s", wpsRelDir, "geogrid.detail.log geogrid.log.*")
	server.ExecRetry(fmt.Sprintf("mpiexec %s -n %d ./geogrid.exe", conf.Values.MpiOptions, conf.Values.GeogridProcCount), wpsPath, "geogrid.detail.log", "{geogrid.detail.log,geogrid.log.????}")
	logFile := join(wpsPath, "geogrid.log.0000")
	logf := errors.CheckResult(os.Open(logFile))
	defer logf.Close()

	prgs := wrfprocs.ShowGeogridProgress(logf, time.Time{}, time.Time{}.Add(time.Hour))

	var p wrfprocs.Progress
	var endLineFound bool
	for p = range prgs {
		if p.Completed {
			endLineFound = true
			if p.Err != nil {
				errors.FailF("geogrid process failed: %w", p.Err)
			} else {
				log.Info("  - Geogrid process completed successfully.")
			}
		}
	}
	if !endLineFound {
		log.Warning("log file %s is malformed: completion line not found.", logFile)
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
	logFile := join(wpsPath, "ungrib.log")
	logf := errors.CheckResult(os.Open(logFile))
	defer logf.Close()

	prgs := wrfprocs.ShowUngribProgress(logf, time.Time{}, time.Time{}.Add(time.Hour))

	var p wrfprocs.Progress
	var endLineFound bool
	for p = range prgs {
		if p.Completed {
			endLineFound = true
			if p.Err != nil {
				errors.FailF("ungrib process failed: %w", p.Err)
			} else {
				log.Info("  - Ungrib process completed successfully.")
			}
		}
	}
	if !endLineFound {
		log.Warning("log file %s is malformed: completion line not found.", logFile)
	}
}

func (s Simulation) RunMetgrid() {
	wpsPath := folders.WPSProcWorkdir(s.Workdir)
	wpsRelDir := errors.CheckResult(filepath.Rel(s.Workdir, wpsPath))

	log.Info("Running metgrid.\t\t\tDIR: $WORKDIR/%s LOGS: %s", wpsRelDir, "metgrid.detail.log metgrid.log.*")
	server.ExecRetry(fmt.Sprintf("mpiexec %s -n %d ./metgrid.exe", conf.Values.MpiOptions, conf.Values.MetgridProcCount), wpsPath, "metgrid.detail.log", "{metgrid.detail.log,metgrid.log.????}")
	logFile := join(wpsPath, "metgrid.log.0000")
	logf := errors.CheckResult(os.Open(logFile))
	defer logf.Close()

	prgs := wrfprocs.ShowMetgridProgress(logf, time.Time{}, time.Time{}.Add(time.Hour))

	var p wrfprocs.Progress
	var endLineFound bool
	for p = range prgs {
		if p.Completed {
			endLineFound = true
			if p.Err != nil {
				errors.FailF("metgrid process failed: %w", p.Err)
			} else {
				log.Info("  - Metgrid process completed successfully.")
			}
		}
	}
	if !endLineFound {
		log.Warning("log file %s is malformed: completion line not found.", logFile)
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

	logFile := join(wpsPath, "rsl.out.0000")
	logf := errors.CheckResult(os.Open(logFile))
	defer logf.Close()

	prgs := wrfprocs.ShowRealProgress(logf, time.Time{}, time.Time{}.Add(time.Hour))

	var p wrfprocs.Progress
	var endLineFound bool
	for p = range prgs {
		if p.Completed {
			endLineFound = true
			if p.Err != nil {
				errors.FailF("real process failed: %w", p.Err)
			} else {
				log.Info("  - Real process completed successfully.")
			}
		}
	}
	if !endLineFound {
		log.Warning("log file %s is malformed: completion line not found.", logFile)
	}

}

func (s Simulation) RunDa(startTime time.Time, domain int) {

	pathDA := folders.DAProcWorkdir(s.Workdir, startTime, domain)

	daRelDir := errors.CheckResult(filepath.Rel(s.Workdir, pathDA))
	log.Info("Running da_wrfvar for %02d:00 (domain %d)\t\tDIR: $WORKDIR/%s LOGS: %s", startTime.Hour(), domain, daRelDir, "da_wrfvar.detail.log rsl.out.* rsl.error.*")

	server.ExecRetry(fmt.Sprintf("mpirun %s -n %d ./da_wrfvar.exe", conf.Values.MpiOptions, conf.Values.WrfdaProcCount), pathDA, "da_wrfvar.detail.log", "{da_wrfvar.detail.log,rsl.out.????,rsl.error.????}")

	logFile := join(pathDA, "rsl.out.0000")
	logf := errors.CheckResult(os.Open(logFile))
	defer logf.Close()

	prgs := wrfprocs.ShowDAProgress(logf)

	var p wrfprocs.Progress
	var endLineFound bool
	for p = range prgs {
		if p.Completed {
			endLineFound = true
			if p.Err != nil {
				errors.FailF("Da_wrfvar process failed: %w", p.Err)
			} else {
				log.Info("  - Da_wrfvar process completed successfully.")
			}
		}
	}
	if !endLineFound {
		log.Warning("log file %s is malformed: completion line not found.", logFile)
	}

}

func (s Simulation) RunWrfEnsemble(startTime time.Time, ensnum int) (err error) {
	defer errors.OnFailuresSet(&err)

	return s.runWrf(startTime, ensnum, conf.Values.WrfProcCount)
}

func (s Simulation) RunWrfStep(startTime time.Time) {
	errors.Check(s.runWrf(startTime, 0, conf.Values.WrfStepProcCount))
}

func (s Simulation) runWrf(startTime time.Time, ensnum int, procCount int) (err error) {
	var workdirPath string
	var descr string
	defer errors.OnFailuresSet(&err)
	if ensnum == 0 {
		workdirPath = folders.WrfControlProcWorkdir(s.Workdir, startTime)
		descr = "control"
	} else {
		workdirPath = folders.WrfEnsembleProcWorkdir(s.Workdir, startTime, ensnum)
		descr = fmt.Sprintf("ensemble n. %d", ensnum)
	}

	wrfRelDir := errors.CheckResult(filepath.Rel(s.Workdir, workdirPath))

	log.Info("Running WRF %s for %02d:00\tDIR: $WORKDIR/%s LOGS: %s", descr, startTime.Hour(), wrfRelDir, "wrf.detail.log rsl.out.* rsl.error.*")
	//--cpu-set 0-15 --bind-to core
	var nodes mpiman.SlurmNodesList

	if conf.Values.EnsembleMembers > 0 {
		var ok bool
		nodes, ok = s.Nodes.FindFreeNodes(int(math.Ceil(float64(procCount) / float64(conf.Values.CoresPerNode))))
		if !ok {
			errors.FailF("Not enough free nodes to run WRF")
		}
	}

	logFile := join(workdirPath, "rsl.out.0000")
	endLineFound := make(chan bool)
	go s.parseProgress(workdirPath, logFile, descr, endLineFound)

	cmd := fmt.Sprintf("mpirun %s %s -n %d ./wrf.exe", conf.Values.MpiOptions, nodes.String(), procCount)
	log.Debug("Running command: %s", cmd)
	server.ExecRetry(cmd, workdirPath, "wrf.detail.log", "{wrf.detail.log,rsl.out.????,rsl.error.????}")
	s.Nodes.Dispose(nodes)

	if !<-endLineFound {
		log.Warning("log file is malformed: completion line not found.")
	}

	return nil
}
func (s Simulation) parseProgress(outputDir, logFile, descr string, endLineFound chan bool) {
	defer errors.OnFailuresDo(func(err errors.RunTimeError) {
		log.Error("Error parsing WRF %s progress: %s", descr, err.Error())
	})
	logf := errors.CheckResult(tailor.OpenFile(logFile, 5*time.Second))
	defer logf.Close()

	prgs := wrfprocs.ShowProgress(logf, time.Time{}, time.Time{}.Add(time.Hour))

	outfLogPath := filepath.Join(s.Workdir, "output_files.log")
	var p wrfprocs.Progress

	for p = range prgs {
		if p.Completed {
			endLineFound <- true
			if p.Err != nil {
				errors.FailF("WRF %s process failed: %w", descr, p.Err)
			} else {
				log.Info("  - WRF %s process completed successfully.", descr)
			}

			outfLog := errors.CheckResult(os.OpenFile(outfLogPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644))
			_, err := outfLog.Write([]byte("COMPLETED\n"))
			outfLog.Close()
			errors.Check(err)

		} else if p.Filename != "" && p.Filename != "restart" {
			outfLog := errors.CheckResult(os.OpenFile(outfLogPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644))
			_, err := outfLog.Write([]byte(filepath.Join(outputDir, p.Filename) + "\n"))
			outfLog.Close()
			errors.Check(err)

			log.Info("File produced by %s: %s", descr, p.Filename)
		}

	}
	close(endLineFound)
}
