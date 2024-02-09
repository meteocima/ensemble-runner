package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/meteocima/ensemble-runner/errors"
	"github.com/parro-it/tailor"
)

func main() {
	defer errors.OnFailuresDo(func(err errors.RunTimeError) {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	})
	ReadConf()
	fmt.Println(Conf.PostprocRules)
	log := errors.CheckResult(tailor.OpenFile("fixtures/output_files.log", time.Second))
	defer log.Close()

	scan := bufio.NewScanner(log)
	for scan.Scan() {
		line := scan.Text()
		if line == "COMPLETED" {
			break
		}
		file := filepath.Base(line)
		for rule, cmd := range Conf.PostprocRules {
			if errors.CheckResult(regexp.Match(rule, []byte(file))) {
				fmt.Printf("%s matched for %s\n", rule, file)
				break
			}
			_ = cmd
		}
		//fmt.Printf("%s\n", file)
	}

	errors.Check(scan.Err())

	//buf := make([]byte, 1024)
	//n := errors.CheckResult(log.Read(buf))
	//fmt.Println("Hello, playground: ", string(buf[:n]))
}
