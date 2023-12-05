package wrfprocs

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"
)

type UngribParser struct {
	R       io.Reader
	Curr    UngribLineInfo
	Err     error
	scanner *bufio.Scanner
}
type UngribLineType int

const (
	UngribInventoryLine UngribLineType = iota
	UngribReprocessLine
	UngribSuccessLine
)

type UngribLineInfo struct {
	Type UngribLineType
	Dt   time.Time
}

var reUngribReprocess = regexp.MustCompile(`First pass done, doing a reprocess`)
var reUngribInventory = regexp.MustCompile(`Inventory for date = (?P<Dt>.+)`)

func ShowUngribProgress(r io.Reader, start, end time.Time) chan Progress {

	ch := make(chan Progress)
	go func() {
		p := UngribParser{R: r}
		defer close(ch)

		reprocess := 0
		duration := end.Sub(start)
		lastProgress := 0
		for p.Read() {
			if p.Curr.Type == UngribInventoryLine {
				currDuration := p.Curr.Dt.Sub(start)
				currProgress := reprocess + int(currDuration*50/duration)
				if currProgress != lastProgress {
					lastProgress = currProgress
					ch <- Progress{Val: currProgress}

				}
				continue

			}
			if p.Curr.Type == UngribReprocessLine {
				reprocess = 50
				continue
			}

			if p.Curr.Type == UngribSuccessLine {
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

func (p *UngribParser) Read() bool {
	if p.scanner == nil {
		p.scanner = bufio.NewScanner(p.R)
	}
	for {
		if !p.scanner.Scan() {
			p.Err = p.scanner.Err()

			return false
		}

		line := p.scanner.Text()

		if groups := reUngribInventory.FindStringSubmatch(line); groups != nil {
			if len(groups) < 2 {
				p.Err = fmt.Errorf("malformed inventory line `%s`", line)
				return false
			}
			p.Curr.Type = UngribInventoryLine
			if p.Curr.Dt, p.Err = time.Parse("2006-01-02 15:04:05", groups[1]); p.Err != nil {
				return false
			}
			return true
		}

		if reUngribReprocess.MatchString(line) {
			p.Curr.Type = UngribReprocessLine
			p.Curr.Dt = time.Time{}
			return true
		}

		if strings.Contains(line, "Successful completion of program ungrib.exe") {
			p.Curr.Type = UngribSuccessLine
			p.Curr.Dt = time.Time{}
			return true
		}

	}
}
