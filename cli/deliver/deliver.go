package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/meteocima/ensemble-runner/errors"
	"github.com/meteocima/ensemble-runner/folders"
	"github.com/meteocima/ensemble-runner/log"
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
	//fmt.Printf("postprocd: %s\n", postprocd)

	postprocdFile := errors.CheckResult(tailor.OpenFile(postprocd, time.Second*30))
	defer postprocdFile.Close()
	scan := bufio.NewScanner(postprocdFile)

	for scan.Scan() {
		line := scan.Bytes()
		var ppc PostProcessCompleted
		errors.Check(json.Unmarshal(line, &ppc))

		if ppc.Kind == RawAuxFile && ppc.Domain == 3 {
			cmd := fmt.Sprintf("scp %s del-continuum:/home/silvestro/Flood_Proofs_Italia2p0/MeteoModel/WrfOL/%s", ppc.FilePath, filepath.Base(ppc.FilePath))
			//server.ExecRetry(cmd, workDir, "deliv-continuum.log", "deliv-continuum.log")
			fmt.Println(cmd)
		} else if ppc.Kind == WrfOutFile && ppc.Domain == 3 {

			fileInst := startInstant.Add(time.Duration(ppc.ProgrHour) * time.Hour)
			fileInstS := fileInst.Format("2006010215")
			filename := fmt.Sprintf("wrfcima_%s-%d.grb2", fileInstS, ppc.ProgrHour)

			// delivery AWS
			cmd := fmt.Sprintf("scp %s del-repo:/share/wrf_repository/%s", ppc.FilePath, filename)
			//server.ExecRetry(cmd, workDir, "deliv-aws.log", "deliv-aws.log")
			fmt.Println(cmd)
			// delivery VdA
			cmd = fmt.Sprintf("scp %s del-vda:/home/WRF/%s", ppc.FilePath, filename)
			//server.ExecRetry(cmd, workDir, "deliv-vda.log", "deliv-vda.log")
			fmt.Println(cmd)
			// delivery arpal
			cmd = fmt.Sprintf("echo put %s /cima2lig/WRF/%s | sftp del-arpal", ppc.FilePath, filename)
			//server.ExecRetry(cmd, workDir, "deliv-arpal.log", "deliv-arpal.log")
			fmt.Println(cmd)

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
				fmt.Fprintf(phaseF, "wrfcima_%s-%d.grb2\n", startInstant.Format("2006010215"), i)
			}
			phaseF.Close()

			cmd := fmt.Sprintf("scp %s del-repo:/share/wrf_repository/%s", phaseFname, filepath.Base(phaseFname))
			//server.ExecRetry(cmd, workDir, "deliv-aws.log", "deliv-aws.log")
			fmt.Println(cmd)
			os.Remove(phaseFname)

		} else if ppc.Kind == Completed {
			fileInstS := startInstant.Format("2006-01-02-15")
			filename := fmt.Sprintf("regr-d01-%s.nc", fileInstS)
			targetDir := fmt.Sprintf("/wrf-world/Native/%04d/%02d/%02d/%04d", startInstant.Year(), startInstant.Month(), startInstant.Day(), startInstant.Hour())
			targetName := fmt.Sprintf("rg_wrf_d01-%s_00UTC.nc", startInstant.Format("2006010215"))
			cmd := fmt.Sprintf("scp %s drihm:%s", filepath.Join(workDir, "results/aux", filename), filepath.Join(targetDir, targetName))
			//server.ExecRetry(cmd, workDir, "deliv-dewetra-d01.log", "deliv-dewetra-d01.log")
			fmt.Println(cmd)

			filename = fmt.Sprintf("regr-d03-%s.nc", fileInstS)
			targetDir = fmt.Sprintf("/share/archivio/experience/data/MeteoModels/WRF_ARPAL/%04d/%02d/%02d/%04d", startInstant.Year(), startInstant.Month(), startInstant.Day(), startInstant.Hour())
			targetName = fmt.Sprintf("rg_wrf-%s_00UTC.nc", startInstant.Format("2006010215"))
			cmd = fmt.Sprintf("scp %s del-dewetra:%s", filepath.Join(workDir, "results/aux", filename), filepath.Join(targetDir, targetName))
			//server.ExecRetry(cmd, workDir, "deliv-dewetra-d01.log", "deliv-dewetra-d01.log")
			fmt.Println(cmd)
			break
		}
	}

	errors.Check(scan.Err())
}
