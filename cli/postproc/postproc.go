package main

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/meteocima/ensemble-runner/errors"
	"github.com/meteocima/ensemble-runner/folders"
	"github.com/meteocima/ensemble-runner/log"
	"github.com/meteocima/ensemble-runner/server"
	"github.com/meteocima/ensemble-runner/simulation"
	"github.com/parro-it/tailor"
)

func main() {
	defer errors.OnFailuresDo(func(err errors.RunTimeError) {
		log.Error("Error: %s\n", err)
		os.Exit(1)
	})
	ReadConf()
	folders.Initialize()

	startInstant := errors.CheckResult(time.Parse(
		simulation.ShortDtFormat,
		os.Getenv("START_FORECAST"),
	))
	outfile := filepath.Join(simulation.Workdir(startInstant), "output_files.log")
	outlog := errors.CheckResult(tailor.OpenFile(outfile, time.Second))
	defer outlog.Close()

	domainRe := errors.CheckResult(regexp.Compile(`_d(\d\d)_`))
	instantRe := errors.CheckResult(regexp.Compile(`\d\d\d\d-\d\d-\d\d_\d\d:\d\d:\d\d`))

	scan := bufio.NewScanner(outlog)
	maxConcurrent := make(chan struct{}, 5)
	var allDone sync.WaitGroup
	var failedLock sync.Mutex
	var failed []string

	for scan.Scan() {
		line := scan.Text()
		if line == "COMPLETED" {
			break
		}

		var command string
		for rule, cmd := range Conf.PostprocRules {
			file := filepath.Base(line)

			if !errors.CheckResult(regexp.Match(rule, []byte(file))) {
				continue
			}
			allDone.Add(1)
			log.Info("Postprocess enqueued for %s\n", file)
			command = cmd
			break
		}

		if command == "" {
			log.Info("No postprocess rule found for %s\n", line)
			continue
		}

		go func(line string, cmd string) {
			maxConcurrent <- struct{}{}
			file := filepath.Base(line)
			defer errors.OnFailuresDo(func(err errors.RunTimeError) {
				log.Error("Error: %s\n", err)
				failedLock.Lock()
				failed = append(failed, file)
				failedLock.Unlock()
			})

			chunk := domainRe.FindSubmatch([]byte(file))[1]
			var domain string
			if len(chunk) == 0 {
				domain = "0"
			} else {
				domain = string(chunk)
			}

			for len(domain) > 1 && domain[0] == '0' {
				domain = domain[1:]
			}

			chunk = instantRe.Find([]byte(file))
			var instant string
			if len(chunk) == 0 {
				instant = ""
			} else {
				instant = string(chunk)
			}

			defer func() { <-maxConcurrent }()
			log.Info("Running `%s` for %s\n", cmd, file)

			simWorkdir := simulation.Workdir(startInstant)
			server.Exec(cmd, simWorkdir, "",
				"FILE_PATH", line,
				"FILE", file,
				"DIR", filepath.Dir(line),
				"DOMAIN", domain,
				"INSTANT", instant,
				"SIM_WORKDIR", simWorkdir,
			)
			log.Info("Postprocess completed for %s\n", file)
			allDone.Done()
		}(line, command)

	}
	errors.Check(scan.Err())
	allDone.Wait()
	if len(failed) > 0 {
		log.Warning("Postprocess for these files failed: %s\n", strings.Join(failed, ", "))
	}

}
