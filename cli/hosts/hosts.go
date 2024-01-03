package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/meteocima/ensemble-runner/mpiman"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: hosts 'slurm'|'cores'|<hosts string>")
		os.Exit(1)
	}
	if os.Args[1] == "cores" {
		fmt.Printf("%d\n", runtime.NumCPU())
		os.Exit(0)
	}
	var hostsStr string
	if os.Args[1] == "slurm" {
		var ok bool
		hostsStr, ok = os.LookupEnv("SLURM_NODELIST")
		if !ok {
			fmt.Fprintln(os.Stderr, "$SLURM_NODELIST not set")
			os.Exit(1)
		}
	} else {
		hostsStr = os.Args[1]
	}
	hosts, err := mpiman.ParseHosts(hostsStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid hosts string: %s", err)
		os.Exit(1)
	}
	for _, host := range hosts {
		fmt.Println(host)
	}

}
