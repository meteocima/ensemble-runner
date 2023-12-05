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

type MetgridParser struct {
	R       io.Reader
	Curr    MetgridLineInfo
	Err     error
	scanner *bufio.Scanner
}
type MetgridLineType int

const (
	MetgridDomainLine MetgridLineType = iota
	MetgridProcessTimeLine
	MetgridSuccessLine
)

type MetgridLineInfo struct {
	Type      MetgridLineType
	Curr, Tot int64
	Dt        time.Time
}

var reMetgridProcessTime = regexp.MustCompile(`Preparing to process output time (?P<Dt>.+)`)
var reMetgridDomain = regexp.MustCompile(`Processing domain (?P<Curr>\d+) of (?P<Tot>\d+)`)

func ShowMetgridProgress(r io.Reader, start, end time.Time) chan Progress {
	ch := make(chan Progress)
	go func() {
		p := MetgridParser{R: r}
		defer close(ch)

		lastProgress := 0
		currDomain := 0
		totDomain := 0
		duration := end.Sub(start)

		for p.Read() {
			if p.Curr.Type == MetgridDomainLine {
				currDomain = int(p.Curr.Curr) - 1
				totDomain = int(p.Curr.Tot)
				continue
			}
			if p.Curr.Type == MetgridProcessTimeLine {
				currDuration := p.Curr.Dt.Sub(start)
				currProgress := (currDomain * 100 / totDomain) + int(currDuration*100/duration)/totDomain
				if currProgress != lastProgress {
					ch <- Progress{Val: currProgress}
					lastProgress = currProgress
				}
				continue
			}

			if p.Curr.Type == MetgridSuccessLine {
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

func (p *MetgridParser) Read() bool {
	if p.scanner == nil {
		p.scanner = bufio.NewScanner(p.R)
	}
	for {
		if !p.scanner.Scan() {
			p.Err = p.scanner.Err()
			return false
		}

		line := p.scanner.Text()

		if groups := reMetgridProcessTime.FindStringSubmatch(line); groups != nil {
			if len(groups) < 2 {
				p.Err = fmt.Errorf("malformed time process line `%s`", line)
				return false
			}
			p.Curr.Type = MetgridProcessTimeLine
			if p.Curr.Dt, p.Err = time.Parse("2006-01-02_15", groups[1]); p.Err != nil {
				return false
			}
			return true
		}

		if groups := reMetgridDomain.FindStringSubmatch(line); groups != nil {
			if len(groups) < 3 {
				p.Err = fmt.Errorf("malformed domain line `%s`", line)
				return false
			}
			p.Curr.Type = MetgridDomainLine
			if p.Curr.Curr, p.Err = strconv.ParseInt(groups[1], 10, 64); p.Err != nil {
				return false
			}
			if p.Curr.Tot, p.Err = strconv.ParseInt(groups[2], 10, 64); p.Err != nil {
				return false
			}
			return true
		}

		if strings.Contains(line, "Successful completion of program metgrid.exe") {
			p.Curr.Type = MetgridSuccessLine
			p.Curr.Curr = 0
			p.Curr.Tot = 0
			return true
		}

	}
}
