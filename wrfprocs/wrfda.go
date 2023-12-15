package wrfprocs

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type DAParser struct {
	R       io.Reader
	Curr    LineInfo
	Err     error
	scanner *bufio.Scanner
}

func (p *DAParser) Read() bool {
	if p.scanner == nil {
		p.scanner = bufio.NewScanner(p.R)
	}
	for {
		if !p.scanner.Scan() {
			p.Err = p.scanner.Err()
			return false
		}

		line := p.scanner.Text()

		if strings.Contains(line, "WRF-Var completed successfully") {
			p.Curr.Type = SuccessLine
			return true
		}

	}
}

func ShowDAProgress(r io.Reader) chan Progress {
	ch := make(chan Progress)
	go func() {
		p := DAParser{R: r}
		defer close(ch)

		for p.Read() {
			if p.Curr.Type == SuccessLine {
				ch <- Progress{Err: p.Err, Completed: true, Val: 100}
				return
			}
		}

		if p.Err == nil {
			p.Err = fmt.Errorf("`success` line not found")
		}
		ch <- Progress{Err: p.Err, Completed: true}

	}()
	return ch
}
