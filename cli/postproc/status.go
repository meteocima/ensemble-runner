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

type PostProcessStatus struct {
	CompletedCh     <-chan PostProcessCompleted
	SimWorkdir      string
	SimStartInstant time.Time
	OUTDone         [49]bool
	AUXDone         [49]bool
	Done            chan struct{}
}

func (stat *PostProcessStatus) Run() {
	postProcessedFile := filepath.Join(stat.SimWorkdir, "postprocd_files.log")
	postProcd := errors.CheckResult(os.OpenFile(postProcessedFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644))
	defer postProcd.Close()

	for completed := range stat.CompletedCh {
		fmt.Fprintf(postProcd, `{"domain": %d, "progr": %d, "kind": "%s", "file": "%s"}`+"\n", completed.Domain, completed.ProgrHour, completed.Kind.String(), completed.FilePath)
		if completed.Kind != WrfOutFile || completed.Domain != 3 {
			stat.AUXDone[completed.ProgrHour] = true
			stat.checkAllAUXCompleted(completed, postProcd)
			continue
		}
		stat.OUTDone[completed.ProgrHour] = true
		stat.checkPhaseCompleted(completed, postProcd)

	}

	close(stat.Done)
}

func (stat *PostProcessStatus) checkAllAUXCompleted(completed PostProcessCompleted, postProcd *os.File) {
	allDone := true
	for i := 0; i <= 48; i++ {
		if !stat.AUXDone[i] {
			allDone = false
			break
		}
	}
	if allDone {
		script := filepath.Join(folders.Rootdir, "scripts/postproc-aux-end.sh")
		log := "postproc-aux-end.log"
		server.ExecRetry(script, stat.SimWorkdir, log, log,
			"SIM_WORKDIR", stat.SimWorkdir,
			"RUNDATE", stat.SimStartInstant.Format("2006-01-02-15"),
		)
	}
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
	phaseCompleted := true
	for i := firstPhaseHour; i <= lastPhaseHour; i++ {
		if !stat.OUTDone[i] {
			phaseCompleted = false
			break
		}
	}

	if phaseCompleted {
		fmt.Fprintf(postProcd, `{"progr": %d, "kind": "Phase"}`+"\n", phase)
	}
}

func (fk FileKind) String() string {
	switch fk {
	case WrfOutFile:
		return "WrfOutFile"
	case AuxFile:
		return "AuxFile"
	default:
		return "Unknown"
	}
}
