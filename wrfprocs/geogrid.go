package wrfprocs

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type GeogridParser struct {
	R       io.Reader
	Curr    GeogridLineInfo
	Err     error
	scanner *bufio.Scanner
}
type GeogridLineType int

const (
	GeogridDomainLine GeogridLineType = iota
	GeogridFieldLine
	GeogridSuccessLine
)

type GeogridLineInfo struct {
	Type      GeogridLineType
	Curr, Tot int64
}

var reGeogridField = regexp.MustCompile(`Processing field (?P<Curr>\d+) of (?P<Tot>\d+)`)
var reGeogridDomain = regexp.MustCompile(`Processing domain (?P<Curr>\d+) of (?P<Tot>\d+)`)

func ShowGeogridProgress(r io.Reader, start, end time.Time) chan Progress {
	ch := make(chan Progress)
	go func() {
		p := GeogridParser{R: r}
		defer close(ch)

		lastProgress := 0
		currDomain := 0
		totDomain := 0
		for p.Read() {
			if p.Curr.Type == GeogridDomainLine {
				currDomain = int(p.Curr.Curr) - 1
				totDomain = int(p.Curr.Tot)
				continue
			}
			if p.Curr.Type == GeogridFieldLine {
				currProgress := (currDomain * 100 / totDomain) + int(p.Curr.Curr*100/p.Curr.Tot)/totDomain
				if currProgress != lastProgress {
					ch <- Progress{Val: currProgress}
					lastProgress = currProgress
				}
				continue
			}

			if p.Curr.Type == GeogridSuccessLine {
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

func (p *GeogridParser) Read() bool {
	if p.scanner == nil {
		p.scanner = bufio.NewScanner(p.R)
	}
	for {
		if !p.scanner.Scan() {
			p.Err = p.scanner.Err()
			return false
		}

		line := p.scanner.Text()

		if groups := reGeogridField.FindStringSubmatch(line); groups != nil {
			if len(groups) < 3 {
				p.Err = fmt.Errorf("malformed field line `%s`", line)
				return false
			}
			p.Curr.Type = GeogridFieldLine
			if p.Curr.Curr, p.Err = strconv.ParseInt(groups[1], 10, 64); p.Err != nil {
				return false
			}
			if p.Curr.Tot, p.Err = strconv.ParseInt(groups[2], 10, 64); p.Err != nil {
				return false
			}
			return true
		}

		if groups := reGeogridDomain.FindStringSubmatch(line); groups != nil {
			if len(groups) < 3 {
				p.Err = fmt.Errorf("malformed domain line `%s`", line)
				return false
			}
			p.Curr.Type = GeogridDomainLine
			if p.Curr.Curr, p.Err = strconv.ParseInt(groups[1], 10, 64); p.Err != nil {
				return false
			}
			if p.Curr.Tot, p.Err = strconv.ParseInt(groups[2], 10, 64); p.Err != nil {
				return false
			}
			return true
		}

		if strings.Contains(line, "Successful completion of program geogrid.exe") {
			p.Curr.Type = GeogridSuccessLine
			p.Curr.Curr = 0
			p.Curr.Tot = 0
			return true
		}

	}
}
