package main

import (
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
		log.Error("Simulation failed: %s\n", err)
		os.Exit(1)
	})

	folders.Initialize(false)
	conf.Initialize()
	log.SetLevel(log.LevelDebug)

	if _, ok := os.LookupEnv("START_FORECAST"); ok {
		simulation.RunForecastFromEnv()
	} else {
		simulation.RunForecastsFromInputs()
	}

}
