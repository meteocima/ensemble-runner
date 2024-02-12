package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/meteocima/ensemble-runner/errors"
	"github.com/meteocima/ensemble-runner/folders"
	"github.com/meteocima/ensemble-runner/server"
	"github.com/meteocima/ensemble-runner/simulation"
	"github.com/parro-it/tailor"
)

func main() {
	defer errors.OnFailuresDo(func(err errors.RunTimeError) {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	})
	ReadConf()
	folders.Initialize()

	startInstant := errors.CheckResult(time.Parse(
		simulation.ShortDtFormat,
		os.Getenv("START_FORECAST"),
	))
	outfile := filepath.Join(simulation.Workdir(startInstant), "output_files.log")
	log := errors.CheckResult(tailor.OpenFile(outfile, time.Second))
	defer log.Close()

	domainRe := errors.CheckResult(regexp.Compile(`_d(\d\d)_`))
	instantRe := errors.CheckResult(regexp.Compile(`\d\d\d\d-\d\d-\d\d_\d\d:\d\d:\d\d`))

	scan := bufio.NewScanner(log)
	maxConcurrent := make(chan struct{}, 5)
	var allDone sync.WaitGroup
	var failedLock sync.Mutex
	var failed []string

	for scan.Scan() {
		line := scan.Text()
		if line == "COMPLETED" {
			break
		}

		for rule, cmd := range Conf.PostprocRules {
			file := filepath.Base(line)

			if errors.CheckResult(regexp.Match(rule, []byte(file))) {
				allDone.Add(1)
				fmt.Printf("Postprocess enqueued for %s\n", file)

				go func(line string, cmd string) {
					maxConcurrent <- struct{}{}
					file := filepath.Base(line)
					defer errors.OnFailuresDo(func(err errors.RunTimeError) {
						fmt.Fprintf(os.Stderr, "Error: %s\n", err)
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
					fmt.Printf("Running `%s` for %s\n", cmd, file)

					server.Exec(cmd, simulation.Workdir(startInstant), "",
						"FILE_PATH", line,
						"FILE", file,
						"DIR", filepath.Dir(line),
						"DOMAIN", domain,
						"INSTANT", instant,
					)
					fmt.Printf("Postprocess completed for %s\n", file)
					allDone.Done()
				}(line, cmd)
				break
			}
			_ = cmd
		}
	}
	errors.Check(scan.Err())
	allDone.Wait()
	fmt.Printf("Postprocess for these files failed: %s\n", strings.Join(failed, ", "))

}
