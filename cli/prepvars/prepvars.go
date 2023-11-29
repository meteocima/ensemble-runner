package main

import (
	"fmt"
	"math"
	"os"
	"time"
)

func main() {
	start, err := time.Parse(ShortDtFormat, os.Getenv("START_DATE"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: $START_DATE: %s", err)
		os.Exit(1)
	}
	end, err := time.Parse(ShortDtFormat, os.Getenv("END_DATE"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: $END_DATE: %s", err)
		os.Exit(1)
	}
	createTemplateArgs(start, end)
}

var IsoFormat = "2006-01-02_15:00:00"
var ShortDtFormat = "2006-01-02-15"

func dumpVar(name, val string) {
	w := "export"
	if len(os.Args) > 1 && os.Args[1] == "--local" {
		w = "local"
	}
	fmt.Printf("%s %s=\"%s\"\n", w, name, val)
}

func createTemplateArgs(start, end time.Time) {
	hours := int(math.Round(end.Sub(start).Hours()))

	dumpVar("RUN_HOURS", fmt.Sprintf("%02d", hours))
	dumpVar("START_DAY", fmt.Sprintf("%02d", start.Day()))
	dumpVar("START_MONTH", fmt.Sprintf("%02d", start.Month()))
	dumpVar("START_YEAR", fmt.Sprintf("%04d", start.Year()))
	dumpVar("START_HOUR", fmt.Sprintf("%02d", start.Hour()))
	dumpVar("ANL_DATE", start.Format(IsoFormat))

	dumpVar("WIN_MIN", start.Add(-1*time.Hour).Format(IsoFormat))
	dumpVar("WIN_MAX", start.Add(1*time.Hour).Format(IsoFormat))

	dumpVar("END_DAY", fmt.Sprintf("%02d", end.Day()))
	dumpVar("END_MONTH", fmt.Sprintf("%02d", end.Month()))
	dumpVar("END_YEAR", fmt.Sprintf("%04d", end.Year()))
	dumpVar("END_HOUR", fmt.Sprintf("%02d", end.Hour()))

	metgridLevels := 34
	if start.Before(time.Date(2019, time.June, 12, 12, 0, 0, 0, time.UTC)) {
		metgridLevels = 32
	}
	if start.Before(time.Date(2016, time.May, 11, 12, 0, 0, 0, time.UTC)) {
		metgridLevels = 27
	}
	dumpVar("METGRID_LEVELS", fmt.Sprintf("%d", metgridLevels))

	if hours > 24 {
		dumpVar("METGRID_CONSTANTS", "constants_name = 'TAVGSFC',")
	} else {
		dumpVar("METGRID_CONSTANTS", "")
	}

	var season string
	// we use an approximation
	// to calculate season
	switch start.Month() {
	case 12, 1, 2:
		season = "winter"
	case 3, 4, 5:
		season = "spring"
	case 6, 7, 8:
		season = "summer"
	case 9, 10, 11:
		season = "fall"
	}

	dumpVar("SEASON", season)

}
