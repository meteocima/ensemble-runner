package main

import (
	"bufio"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/meteocima/ensemble-runner/errors"
	"github.com/meteocima/ensemble-runner/log"
	"github.com/meteocima/ensemble-runner/server"
	"github.com/meteocima/ensemble-runner/simulation"
	"github.com/parro-it/tailor"
)

var domainRe = errors.CheckResult(regexp.Compile(`_d(\d\d)_`))
var instantRe = errors.CheckResult(regexp.Compile(`\d\d\d\d-\d\d-\d\d_\d\d:\d\d:\d\d`))

type PostProcessCommand struct {
	FilePath string
	Cmd      string
}

type Worker struct {
	Cmds           <-chan PostProcessCommand
	FilesCompleted chan<- PostProcessCompleted
	SimWorkdir     string
	AllDone        *sync.WaitGroup
	Failures       []PostProcessCommand
	StartInstant   time.Time
	Index          int
}

func (w *Worker) runCommand(ppc PostProcessCommand) {
	file := filepath.Base(ppc.FilePath)
	defer errors.OnFailuresDo(func(err errors.RunTimeError) {
		log.Warning("WORKER %d: postprocess failed for file %s. Will be retried at end. Error: %s", w.Index, filepath.Base(ppc.FilePath), err)
		w.Failures = append(w.Failures, ppc)
	})

	chunk := domainRe.FindSubmatch([]byte(file))[1]
	var domainS string
	domain := errors.CheckResult(strconv.ParseInt(string(chunk), 10, 64))
	domainS = strconv.FormatInt(domain, 10)

	chunk = instantRe.Find([]byte(file))
	var instantS string
	if len(chunk) == 0 {
		instantS = ""
	} else {
		instantS = string(chunk)
	}
	instant := errors.CheckResult(time.Parse("2006-01-02_15:04:05", instantS))

	log.Info("Running postprocessing for file %s", file)
	log.Debug("\t Command for file %s: `%s` ", file, ppc.Cmd)

	server.Exec(ppc.Cmd, w.SimWorkdir, "",
		"FILE_PATH", ppc.FilePath,
		"FILE", file,
		"DIR", filepath.Dir(ppc.FilePath),
		"DOMAIN", domainS,
		"INSTANT", instantS,
		"SIM_WORKDIR", w.SimWorkdir,
	)
	log.Info("Postprocess completed for %s", file)
	progrHour := int(instant.Sub(w.StartInstant).Hours())
	var kind FileKind
	var filePath string
	if strings.HasPrefix(file, "wrfout") {
		kind = WrfOutFile
		filePath = filepath.Join(w.SimWorkdir, fmt.Sprintf("results/out/out_regr_%s.grb", instantS))
	} else if strings.HasPrefix(file, "aux") {
		kind = AuxFile
		filePath = filepath.Join(w.SimWorkdir, fmt.Sprintf("results/aux/aux-regr-d%02d-%s.nc", domain, instantS))

		w.FilesCompleted <- PostProcessCompleted{
			Domain:    int(domain),
			ProgrHour: progrHour,
			Kind:      RawAuxFile,
			FilePath:  filepath.Join(w.SimWorkdir, "results/rawaux", file),
		}
	} else {
		errors.FailF("Unknown file kind for %s", file)
	}
	w.FilesCompleted <- PostProcessCompleted{
		Domain:    int(domain),
		ProgrHour: progrHour,
		Kind:      kind,
		FilePath:  filePath,
	}
}

func (w *Worker) Run() {
	for ppc := range w.Cmds {
		w.runCommand(ppc)
	}
	for i := 1; i <= 5 && len(w.Failures) > 0; i++ {
		log.Info("WORKER %d: Retrying failed processes. Iteration %d", w.Index, i)
		failures := w.Failures
		w.Failures = nil
		for _, ppc := range failures {
			w.runCommand(ppc)
		}
	}

	w.AllDone.Done()
}

func RunPostProcessing(startInstant time.Time, totHours int) {
	simWorkdir := simulation.Workdir(startInstant)
	outfile := filepath.Join(simWorkdir, "output_files.log")

	completedCh := make(chan PostProcessCompleted)
	status := PostProcessStatus{
		CompletedCh:          completedCh,
		SimWorkdir:           simWorkdir,
		SimStartInstant:      startInstant,
		OUTDone:              [49]bool{},
		AUXDoneD1:            [49]bool{},
		AUXDoneD3:            [49]bool{},
		FinalAUXPostProcDone: false,
		OutPhasesDone:        [4]bool{},
		Done:                 make(chan struct{}),
		TotHours:             totHours,
	}
	go status.Run()

	cmds := make(chan PostProcessCommand, 49*6)
	allDone := sync.WaitGroup{}
	allDone.Add(5)
	var workers []*Worker
	for i := 0; i < 5; i++ {
		w := Worker{
			Cmds:           cmds,
			SimWorkdir:     simWorkdir,
			AllDone:        &allDone,
			StartInstant:   startInstant,
			FilesCompleted: completedCh,
			Index:          i,
		}
		workers = append(workers, &w)
		go w.Run()
	}

	outlog := errors.CheckResult(tailor.OpenFile(outfile, time.Second))
	defer outlog.Close()
	scan := bufio.NewScanner(outlog)

	for scan.Scan() {
		line := scan.Text()
		if line == "COMPLETED" {
			break
		}

		var command string
		file := filepath.Base(line)

		for rule, cmd := range Conf.PostprocRules {
			if !errors.CheckResult(regexp.Match(rule, []byte(file))) {
				continue
			}
			command = cmd
			break
		}

		if command == "" {
			log.Debug("No postprocess rule found for %s", filepath.Base(line))
			continue
		}

		cmds <- PostProcessCommand{
			FilePath: line,
			Cmd:      command,
		}
		log.Info("Postprocess enqueued for %s", filepath.Base(line))

	}
	close(cmds)
	allDone.Wait()

	var allFailures []PostProcessCommand

	for _, w := range workers {
		allFailures = append(allFailures, w.Failures...)
	}

	if len(allFailures) > 0 {
		var filesFailed []string
		for _, ppc := range allFailures {
			filesFailed = append(filesFailed, filepath.Base(ppc.FilePath))
		}
		filesFailedS := "\n\t" + strings.Join(filesFailed, "\n\t")
		log.Error("Postprocessing completed, some processes failed after 5 retries. Failed files: %v", filesFailedS)
	} else {
		log.Info("Postprocessing completed, all files successfully postprocessed.")
	}
	close(completedCh)
	<-status.Done
	errors.Check(scan.Err())
}
