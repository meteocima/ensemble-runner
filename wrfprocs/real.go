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

type RealParser struct {
	R       io.Reader
	Curr    RealLineInfo
	Err     error
	scanner *bufio.Scanner
}
type RealLineType int

const (
	RealProcessLine RealLineType = iota
	RealSuccessLine
)

type RealLineInfo struct {
	Type      RealLineType
	Curr, Tot int64
	Dt        time.Time
}

var reRealProcess = regexp.MustCompile(`Domain  1: Current date being processed: (?P<Dt>\S.+), which is loop #\s*(?P<Curr>\d+)\s*out of\s*(?P<Tot>\d+)`)

func ShowRealProgress(r io.Reader, start, end time.Time) chan Progress {
	ch := make(chan Progress)
	go func() {
		p := RealParser{R: r}
		defer close(ch)

		lastProgress := 0
		duration := end.Sub(start)

		for p.Read() {

			if p.Curr.Type == RealProcessLine {
				currDuration := p.Curr.Dt.Sub(start)
				currProgress := int(currDuration * 100 / duration)
				if currProgress != lastProgress {
					ch <- Progress{Val: currProgress}
					lastProgress = currProgress
				}
				continue
			}

			if p.Curr.Type == RealSuccessLine {
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

func (p *RealParser) Read() bool {
	if p.scanner == nil {
		p.scanner = bufio.NewScanner(p.R)
	}
	for {
		if !p.scanner.Scan() {
			p.Err = p.scanner.Err()
			return false
		}

		line := p.scanner.Text()

		if groups := reRealProcess.FindStringSubmatch(line); groups != nil {
			if len(groups) < 4 {
				p.Err = fmt.Errorf("malformed process line `%s`", line)
				return false
			}
			p.Curr.Type = RealProcessLine
			if p.Curr.Dt, p.Err = time.Parse("2006-01-02_15:04:05.0000", groups[1]); p.Err != nil {
				return false
			}
			if p.Curr.Curr, p.Err = strconv.ParseInt(groups[2], 10, 64); p.Err != nil {
				return false
			}
			if p.Curr.Tot, p.Err = strconv.ParseInt(groups[3], 10, 64); p.Err != nil {
				return false
			}
			return true
		}

		if strings.Contains(line, "SUCCESS COMPLETE REAL_EM INIT") {
			p.Curr.Type = RealSuccessLine
			p.Curr.Curr = 0
			p.Curr.Tot = 0
			return true
		}

	}
}
