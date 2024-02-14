package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
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
	WrfOutFile = FileKind(0)
	AuxFile    = FileKind(1)
	Phase      = FileKind(2)
)

type PostProcessCompleted struct {
	Domain    int
	ProgrHour int
	Kind      FileKind
	FilePath  string
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

	postprocdFile := errors.CheckResult(tailor.OpenFile(postprocd, time.Second*30))
	defer postprocdFile.Close()
	scan := bufio.NewScanner(postprocdFile)

	for scan.Scan() {
		line := scan.Bytes()
		var ppc PostProcessCompleted
		errors.Check(json.Unmarshal(line, &ppc))
		if ppc.Kind == WrfOutFile {

			fileInst := startInstant.Add(time.Duration(ppc.ProgrHour) * time.Hour)
			fileInstS := fileInst.Format("2006010215")
			filename := fmt.Sprintf("wrfcima_%s-%d.grb2", fileInstS, ppc.ProgrHour)

			// delivery AWS
			cmd := fmt.Sprintf("scp %s del-repo:/share/wrf_repository/%s", ppc.FilePath, filename)
			server.ExecRetry(cmd, workDir, "deliv-aws.log", "deliv-aws.log")

			// delivery VdA
			cmd = fmt.Sprintf("scp %s del-vda:/home/WRF/%s", ppc.FilePath, filename)
			server.ExecRetry(cmd, workDir, "deliv-vda.log", "deliv-vda.log")

			// delivery arpal
			cmd = fmt.Sprintf("sftp del-arpal <<< put %s /cima2lig/WRF/%s", ppc.FilePath, filename)
			server.ExecRetry(cmd, workDir, "deliv-vda.log", "deliv-vda.log")
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

			cmd := fmt.Sprintf("scp %s del-repo:/share/wrf_repository/%s", ppc.FilePath, phaseFname)
			server.ExecRetry(cmd, workDir, "deliv-aws.log", "deliv-aws.log")
			os.Remove(phaseFname)

		}
	}
}
