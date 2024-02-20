package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/meteocima/ensemble-runner/errors"
	"github.com/meteocima/ensemble-runner/folders"
	"github.com/meteocima/ensemble-runner/log"
	"github.com/meteocima/ensemble-runner/server"
	"github.com/meteocima/ensemble-runner/simulation"
	"github.com/parro-it/tailor"
)

type FileKind int

const (
	WrfOutFile FileKind = iota
	AuxFile
	RawAuxFile
	Phase
	Completed
	Unknown
)

var fileKindNames = []string{
	"WrfOutFile",
	"AuxFile",
	"RawAuxFile",
	"Phase",
	"Completed",
	"Unknown",
}

func (fk FileKind) String() string {
	if fk < 0 || fk > Unknown {
		fk = Unknown
	}
	return fileKindNames[fk]
}

func (fk *FileKind) UnmarshalJSON(data []byte) error {
	//fmt.Println("string(data)", string(data))
	for i, name := range fileKindNames {
		if fmt.Sprintf(`"%s"`, name) == string(data) {
			*fk = FileKind(i)
			return nil
		}
	}
	*fk = Unknown
	return nil
}

type PostProcessCompleted struct {
	Domain    int      `json:"domain"`
	ProgrHour int      `json:"progr"`
	Kind      FileKind `json:"kind"`
	FilePath  string   `json:"file"`
}

func main() {
	defer errors.OnFailuresDo(func(err errors.RunTimeError) {
		log.Error("Error: %s", err)
		os.Exit(1)
	})

	folders.Initialize(true)

	startInstant := errors.CheckResult(time.Parse(
		simulation.ShortDtFormat,
		os.Getenv("START_FORECAST"),
	))

	workDir := simulation.Workdir(startInstant)
	postprocd := workDir + "/postprocd_files.log"
	//log.Info("postprocd: %s\n", postprocd)

	postprocdFile := errors.CheckResult(tailor.OpenFile(postprocd, time.Second*30))
	defer postprocdFile.Close()
	scan := bufio.NewScanner(postprocdFile)

	var chanPPC = make(chan PostProcessCompleted)
	var alldone sync.WaitGroup
	alldone.Add(10)
	for i := 0; i < 10; i++ {
		go func() {

			defer alldone.Done()
			for ppc := range chanPPC {
				deliverFile(ppc, workDir, startInstant)
			}
		}()

	}

	for scan.Scan() {
		line := scan.Bytes()
		var ppc PostProcessCompleted
		errors.Check(json.Unmarshal(line, &ppc))

		chanPPC <- ppc

		if ppc.Kind == Completed {
			break
		}
	}
	close(chanPPC)
	alldone.Wait()
	errors.Check(scan.Err())
}

func deliverFile(ppc PostProcessCompleted, workDir string, startInstant time.Time) {
	defer errors.OnFailuresDo(func(err errors.RunTimeError) {
		log.Error("Error: %s", err)
	})

	if ppc.Kind == RawAuxFile && ppc.Domain == 3 {
		// delivery raw aux files to continuum
		cmd := fmt.Sprintf("scp %s del-continuum:/home/silvestro/Flood_Proofs_Italia2p0/MeteoModel/WrfOL/%s", ppc.FilePath, filepath.Base(ppc.FilePath))
		log.Info("Start delivery file %s to continuum", filepath.Base(ppc.FilePath))
		server.ExecRetry(cmd, workDir, "deliv-continuum.log", "deliv-continuum.log")
		log.Info("Delivered file %s to continuum", filepath.Base(ppc.FilePath))
	} else if ppc.Kind == WrfOutFile && ppc.Domain == 3 {

		fileInstS := startInstant.Format("2006010215")
		filename := fmt.Sprintf("wrfcima_%s-%02d.grb2", fileInstS, ppc.ProgrHour)

		// delivery AWS
		cmd := fmt.Sprintf("scp %s del-repo:/share/wrf_repository/%s", ppc.FilePath, filename)
		log.Info("Start delivery file %s to AWS", filename)
		server.ExecRetry(cmd, workDir, "deliv-aws.log", "deliv-aws.log")
		log.Info("Delivered file %s to AWS", filename)

		// delivery VdA
		cmd = fmt.Sprintf("scp %s del-vda:/home/WRF/%s", ppc.FilePath, filename)
		log.Info("Start delivery file %s to VdA", filename)
		server.ExecRetry(cmd, workDir, "deliv-vda.log", "deliv-vda.log")
		log.Info("Delivered file %s to VdA", filename)

		// delivery arpal
		cmd = fmt.Sprintf("echo put %s /cima2lig/WRF/%s | sftp del-arpal", ppc.FilePath, filename)
		log.Info("Start delivery file %s to ARPAL", filename)
		server.ExecRetry(cmd, workDir, "deliv-arpal.log", "deliv-arpal.log")
		log.Info("Delivered file %s to ARPAL", filename)

	} else if ppc.Kind == Phase {
		var firstPhaseHour int
		var lastPhaseHour int

		if ppc.ProgrHour == 1 {
			firstPhaseHour = 0
		} else {
			firstPhaseHour = (ppc.ProgrHour-1)*12 + 1
		}

		lastPhaseHour = ppc.ProgrHour * 12
		phaseFname := fmt.Sprintf("%s/index%d.txt", os.TempDir(), ppc.ProgrHour-1)
		phaseF := errors.CheckResult(os.OpenFile(phaseFname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644))
		for i := firstPhaseHour; i <= lastPhaseHour; i++ {
			fmt.Fprintf(phaseF, "wrfcima_%s-%02d.grb2\n", startInstant.Format("2006010215"), i)
		}
		phaseF.Close()

		// delivery AWS of phase index
		cmd := fmt.Sprintf("scp %s del-repo:/share/wrf_repository/%s", phaseFname, filepath.Base(phaseFname))
		log.Info("Start delivery file %s to AWS", filepath.Base(phaseFname))
		server.ExecRetry(cmd, workDir, "deliv-aws.log", "deliv-aws.log")
		log.Info("Delivered file %s to AWS", filepath.Base(phaseFname))
		os.Remove(phaseFname)
	} else if ppc.Kind == AuxFile && ppc.Domain == 3 {
		// delivery domain 3 to dhrim for final merge and delivery to dewetra and AWS

		targetDir := fmt.Sprintf("/share/ol_leo/%s", startInstant.Format("2006-01-02-15"))
		log.Info("Start delivery file %s to drihm", filepath.Base(ppc.FilePath))
		server.ExecRetry("ssh drihm mkdir -p "+targetDir, workDir, "", "")

		cmd := fmt.Sprintf("scp %s drihm:%s/%s", ppc.FilePath, targetDir, filepath.Base(ppc.FilePath))
		server.ExecRetry(cmd, workDir, "deliv-dewetra-d01.log", "deliv-dewetra-d01.log")
		log.Info("Delivered file %s to Dewetra", filepath.Base(ppc.FilePath))

	} else if ppc.Kind == Completed {
		// delivery domain 1 to Dewetra
		fileInstS := startInstant.Format("2006-01-02-15")
		filename := fmt.Sprintf("regr-d01-%s.nc", fileInstS)
		targetDir := fmt.Sprintf("/wrf-world/Native/%04d/%02d/%02d/%04d", startInstant.Year(), startInstant.Month(), startInstant.Day(), startInstant.Hour())
		targetName := fmt.Sprintf("rg_wrf_d01-%s_00UTC.nc", startInstant.Format("2006010215"))

		log.Info("Start delivery file %s to Dewetra World", targetName)
		server.ExecRetry("ssh del-dewetra-world mkdir -p "+targetDir, workDir, "deliv-dewetra-d01.log", "deliv-dewetra-d01.log")

		cmd := fmt.Sprintf("scp %s del-dewetra-world:%s", filepath.Join(workDir, "results/aux", filename), filepath.Join(targetDir, targetName))
		server.ExecRetry(cmd, workDir, "deliv-dewetra-d01.log", "deliv-dewetra-d01.log")
		log.Info("Delivered file %s to Dewetra World", targetName)

		// delivery domain 3 to Dewetra
		filename = fmt.Sprintf("regr-d03-%s.nc", fileInstS)
		targetDir = fmt.Sprintf("/share/archivio/experience/data/MeteoModels/WRF_ARPAL/%04d/%02d/%02d/%04d", startInstant.Year(), startInstant.Month(), startInstant.Day(), startInstant.Hour())
		targetName = fmt.Sprintf("rg_wrf-%s_00UTC.nc", startInstant.Format("2006010215"))
		log.Info("Start delivery file %s to Dewetra", targetName)
		server.ExecRetry("ssh del-dewetra mkdir -p "+targetDir, workDir, "", "")

		cmd = fmt.Sprintf("scp %s del-dewetra:%s", filepath.Join(workDir, "results/aux", filename), filepath.Join(targetDir, targetName))
		server.ExecRetry(cmd, workDir, "deliv-dewetra-d01.log", "deliv-dewetra-d01.log")
		log.Info("Delivered file %s to Dewetra", targetName)

	}
}
