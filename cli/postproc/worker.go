package main

import (
	"bufio"
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
}

func (w *Worker) runCommand(ppc PostProcessCommand) {
	file := filepath.Base(ppc.FilePath)
	defer errors.OnFailuresDo(func(err errors.RunTimeError) {
		log.Warning("Postprocess failed for file %w. Will be retried at end. Error: %s", filepath.Base(ppc.FilePath), err)
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

	log.Debug("Running `%s` for %s", ppc.Cmd, file)

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
	if strings.HasPrefix(file, "wrfout") {
		kind = WrfOutFile
	} else if strings.HasPrefix(file, "aux") {
		kind = AuxFile
	} else {
		errors.FailF("Unknown file kind for %s", file)
	}
	w.FilesCompleted <- PostProcessCompleted{
		Domain:    int(domain),
		ProgrHour: progrHour,
		Kind:      kind,
	}
}

func (w *Worker) Run() {
	for ppc := range w.Cmds {
		w.runCommand(ppc)
	}
	for i := 1; i <= 5; i++ {
		if len(w.Failures) == 0 {
			break
		}
		log.Info("Retrying failed postprocesses. Iteration %d", i)
		for _, ppc := range w.Failures {
			w.runCommand(ppc)
		}
	}
	w.AllDone.Done()
}

func RunPostProcessing(startInstant time.Time) {
	completedCh := make(chan PostProcessCompleted)

	simWorkdir := simulation.Workdir(startInstant)
	cmds := make(chan PostProcessCommand, 49*6)
	allDone := sync.WaitGroup{}
	allDone.Add(5)
	for i := 0; i < 5; i++ {
		w := Worker{
			Cmds:           cmds,
			SimWorkdir:     simWorkdir,
			AllDone:        &allDone,
			StartInstant:   startInstant,
			FilesCompleted: completedCh,
		}
		go w.Run()
	}

	outfile := filepath.Join(simWorkdir, "output_files.log")
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
			log.Info("No postprocess rule found for %s", filepath.Base(line))
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
	errors.Check(scan.Err())
}
