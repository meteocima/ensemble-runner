package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/meteocima/ensemble-runner/dirprep"
	"golang.org/x/exp/maps"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: dirprep [--print-vars] [--strict] <SRCDIR...> <DSTDIR>\n")
		os.Exit(1)
	}

	var srcdirs []string
	var dstdir string
	var printVars bool
	var strict bool
	var resolvedVars = map[string]struct{}{}
	for _, arg := range os.Args[1 : len(os.Args)-1] {
		if arg == "--print-vars" {
			printVars = true
			continue
		}
		if arg == "--strict" {
			strict = true
			continue
		}
		srcdirs = append(srcdirs, arg)
	}
	dstdir = os.Args[len(os.Args)-1]
	xp := os.Getenv
	if printVars {
		prevXp := xp
		xp = func(key string) string {
			resolvedVars[key] = struct{}{}
			return prevXp(key)
		}
	}
	if strict {
		prevXp := xp
		xp = func(key string) string {
			if _, ok := os.LookupEnv(key); !ok {
				fmt.Fprintf(os.Stderr, "Env variable not found: $%s\n", key)
				os.Exit(1)
			}
			val := prevXp(key)

			return val
		}
	}
	for _, srcdir := range srcdirs {
		err := dirprep.RenderDirEnv(srcdir, dstdir, xp)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
	}
	if printVars {
		names := maps.Keys(resolvedVars)
		sort.Strings(names)
		for _, name := range names {
			fmt.Printf("$%s\n", name)
		}
	}
}
