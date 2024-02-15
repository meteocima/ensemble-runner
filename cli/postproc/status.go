package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/meteocima/ensemble-runner/errors"
	"github.com/meteocima/ensemble-runner/folders"
	"github.com/meteocima/ensemble-runner/server"
)

type FileKind int

const (
	WrfOutFile FileKind = iota
	AuxFile
	RawAuxFile
	Unknown
)

var fileKindNames = []string{
	"WrfOutFile",
	"AuxFile",
	"RawAuxFile",
	"Phase",
	"Unknown",
}

func (fk FileKind) String() string {
	if fk < 0 || fk > Unknown {
		fk = Unknown
	}
	return fileKindNames[fk]
}

type PostProcessCompleted struct {
	Domain    int
	ProgrHour int
	Kind      FileKind
	FilePath  string
}

type PostProcessStatus struct {
	CompletedCh          <-chan PostProcessCompleted
	SimWorkdir           string
	SimStartInstant      time.Time
	OUTDone              [49]bool
	AUXDoneD1            [49]bool
	AUXDoneD3            [49]bool
	FinalAUXPostProcDone bool
	OutPhasesDone        [4]bool
	Done                 chan struct{}
	TotHours             int
}

func (stat *PostProcessStatus) Run() {
	postProcessedFile := filepath.Join(stat.SimWorkdir, "postprocd_files.log")
	postProcd := errors.CheckResult(os.OpenFile(postProcessedFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644))
	defer postProcd.Close()

	for completed := range stat.CompletedCh {
		fmt.Fprintf(postProcd, `{"domain": %d, "progr": %d, "kind": "%s", "file": "%s"}`+"\n", completed.Domain, completed.ProgrHour, completed.Kind.String(), completed.FilePath)
		if completed.Kind == AuxFile {
			if completed.Domain == 1 {
				stat.AUXDoneD1[completed.ProgrHour] = true
			} else if completed.Domain == 3 {
				stat.AUXDoneD3[completed.ProgrHour] = true
			} else {
				errors.FailF("Unknown domain for AUX file %s: %d", completed.FilePath, completed.Domain)
			}

			stat.checkAllAUXCompleted(completed, postProcd)
		} else if completed.Kind == WrfOutFile {
			stat.OUTDone[completed.ProgrHour] = true
			stat.checkPhaseCompleted(completed, postProcd)
		}

		stat.checkAllPostProcessingCompleted(postProcd)

	}

	close(stat.Done)
}

func (stat *PostProcessStatus) checkAllPostProcessingCompleted(postProcd *os.File) {

	if !stat.FinalAUXPostProcDone {
		return
	}

	for i := 0; i < stat.TotHours/12; i++ {
		if !stat.OutPhasesDone[i] {
			return
		}
	}

	fmt.Fprintf(postProcd, `{"kind": "Completed"}`+"\n")

}
func (stat *PostProcessStatus) checkAllAUXCompleted(completed PostProcessCompleted, postProcd *os.File) {
	for i := 0; i <= stat.TotHours; i++ {
		if !stat.AUXDoneD1[i] {
			return
		}

		if !stat.AUXDoneD3[i] {
			return
		}
	}

	script := filepath.Join(folders.Rootdir, "scripts/postproc-aux-end.sh")
	log := "postproc-aux-end.log"
	server.ExecRetry(script, stat.SimWorkdir, log, log,
		"SIM_WORKDIR", stat.SimWorkdir,
		"RUNDATE", stat.SimStartInstant.Format("2006-01-02-15"),
	)
	stat.FinalAUXPostProcDone = true
}

func (stat *PostProcessStatus) checkPhaseCompleted(completed PostProcessCompleted, postProcd *os.File) {
	var phase int
	var firstPhaseHour int
	var lastPhaseHour int

	if completed.ProgrHour == 0 {
		phase = 1
		firstPhaseHour = 0
	} else {
		phase = int(math.Ceil(float64(completed.ProgrHour) / 12))
		firstPhaseHour = (phase-1)*12 + 1
	}

	lastPhaseHour = phase * 12
	for i := firstPhaseHour; i <= lastPhaseHour; i++ {
		if !stat.OUTDone[i] {
			return
		}
	}

	fmt.Fprintf(postProcd, `{"progr": %d, "kind": "Phase"}`+"\n", phase)
	stat.OutPhasesDone[phase-1] = true
}
