package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"

	"github.com/meteocima/ensemble-runner/errors"
)

type FileKind int

const (
	WrfOutFile = FileKind(0)
	AuxFile    = FileKind(1)
)

type PostProcessCompleted struct {
	Domain    int
	ProgrHour int
	Kind      FileKind
}

type PostProcessStatus struct {
	CompletedCh <-chan PostProcessCompleted
	SimWorkdir  string
	OUTDone     [49]bool
	Done        chan struct{}
}

func (stat *PostProcessStatus) Run() {
	postProcessedFile := filepath.Join(stat.SimWorkdir, "postprocd_files.log")
	postProcd := errors.CheckResult(os.OpenFile(postProcessedFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644))
	defer postProcd.Close()

	for completed := range stat.CompletedCh {
		fmt.Fprintf(postProcd, `{"domain": %d, "progr": %d, "kind": "%s"}`+"\n", completed.Domain, completed.ProgrHour, completed.Kind.String())
		if completed.Kind != WrfOutFile || completed.Domain != 3 {
			continue
		}
		stat.OUTDone[completed.ProgrHour] = true

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
			fmt.Fprintf(postProcd, `{"progr": %d, "kind": "phase"}`+"\n", phase)
		}

	}

	close(stat.Done)
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
