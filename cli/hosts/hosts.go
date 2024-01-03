package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"

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
		if e, ok := err.(mpiman.ParseError); ok {
			msg := fmt.Sprintf("Invalid hosts string at character %d: %s.\n", e.Pos, e.Msg)
			fmt.Fprint(os.Stderr, msg)
			fmt.Fprintf(os.Stderr, "%*s╭", e.Pos, " ")
			fmt.Fprintf(os.Stderr, "%s╯\n", strings.Repeat("─", len(msg)-e.Pos-3))
			fmt.Fprintf(os.Stderr, "%*s🡣\n", e.Pos, " ")
			fmt.Fprintf(os.Stderr, "%s\n", e.Src)

		} else {
			fmt.Printf("Invalid hosts string at character: %s.\n", err)
		}
		os.Exit(1)
	}
	for _, host := range hosts {
		fmt.Println(host)
	}

}
