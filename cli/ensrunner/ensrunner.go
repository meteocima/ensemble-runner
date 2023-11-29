package main

import (
	"fmt"
	"os"

	"github.com/meteocima/ensemble-runner/conf"
	"github.com/meteocima/ensemble-runner/errors"
	"github.com/meteocima/ensemble-runner/folders"
	"github.com/meteocima/ensemble-runner/log"
	"github.com/meteocima/ensemble-runner/simulation"
)

func main() {
	log.Info("WRF runner starting. Checking configuration...")

	defer errors.OnFailuresDo(func(err errors.RunTimeError) {
		fmt.Fprintf(os.Stderr, "wrfda ita failed: %s\n", err)
		os.Exit(1)
	})

	folders.Initialize()
	conf.Initialize()
	log.SetLevel(log.LevelDebug)

	sim := simulation.New()

	sim.Run()

}
