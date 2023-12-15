package wrfprocs

import (
	"io"
	"time"
)

/*
type Parser struct {
	R       io.Reader
	Curr    LineInfo
	Err     error
	scanner *bufio.Scanner
}
type LineType int

const (
	CalcLine LineType = iota
	FileOutLine
	FileInputLine
	QuiltingLine
	NtasksLine
	SuccessLine
)

type LineInfo struct {
	Type     LineType
	Instant  time.Time
	Timestep float64
	Domain   int64
	Duration time.Duration
	Filename string
	X, Y     int64
}

var reCalcNoTimestep = regexp.MustCompile(`Timing for main: time (?P<Instant>\d{4}-\d{2}-\d{2}_\d{2}:\d{2}:\d{2}) on domain +(?P<DOM>\d+): +(?P<DUR>[\d|\.]+) elapsed seconds`)
var reCalc = regexp.MustCompile(`Timing for main \(dt= *(?P<Timestep>[\d|\.]+)\): time (?P<Instant>\d{4}-\d{2}-\d{2}_\d{2}:\d{2}:\d{2}) on domain +(?P<DOM>\d+): +(?P<DUR>[\d|\.]+) elapsed seconds`)
var reIO = regexp.MustCompile(`Timing for Writing (?P<File>\S+|filter output) for domain +(?P<DOM>\d+): +(?P<DUR>[\d|\.]+) elapsed seconds`)

func (p *Parser) Read() bool {
	if p.scanner == nil {
		p.scanner = bufio.NewScanner(p.R)
	}
	for {
		if !p.scanner.Scan() {
			p.Err = p.scanner.Err()
			return false
		}

		line := p.scanner.Text()
		if strings.HasPrefix(line, "Timing for main") {
			return p.parseCalcLine(line)
		}
		if strings.HasPrefix(line, "Timing for Writing") {
			return p.parseOutLine(line)
		}
		if strings.HasPrefix(line, "Timing for processing") {
			return p.parseInpLine(line)
		}

		if strings.Contains(line, "wrf: SUCCESS COMPLETE WRF") {
			return p.parseSuccessLine(line)
		}

	}
}

func (p *Parser) parseOutLine(line string) bool {
	groups := reIO.FindStringSubmatch(line)
	if len(groups) < 4 {
		p.Err = fmt.Errorf("malformed I/O line `%s`", line)
		return false
	}
	file := groups[1]
	domain := groups[2]
	duration := groups[3]
	if p.Curr.Domain, p.Err = strconv.ParseInt(domain, 10, 64); p.Err != nil {
		return false
	}
	if p.Curr.Duration, p.Err = time.ParseDuration(duration + "s"); p.Err != nil {
		return false
	}
	p.Curr.Type = FileOutLine
	p.Curr.Timestep = -1
	p.Curr.Instant = time.Time{}
	p.Curr.Filename = file
	return true
}

// Timing for processing wrfinput file (stream 0) for domain        3:    2.18410 elapsed seconds
var reInp = regexp.MustCompile(`Timing for processing (?P<File>.+) for domain +(?P<DOM>\d+): +(?P<DUR>[\d|\.]+) elapsed seconds`)

func (p *Parser) parseSuccessLine(line string) bool {
	p.Curr.Type = SuccessLine
	p.Curr.Timestep = -1
	p.Curr.Instant = time.Time{}
	p.Curr.Filename = ""
	p.Curr.Domain = -1
	p.Curr.Duration = 0
	return true
}

func (p *Parser) parseInpLine(line string) bool {
	groups := reInp.FindStringSubmatch(line)
	if len(groups) < 4 {
		p.Err = fmt.Errorf("malformed I/O line `%s`", line)
		return false
	}
	file := groups[1]
	domain := groups[2]
	duration := groups[3]
	if p.Curr.Domain, p.Err = strconv.ParseInt(domain, 10, 64); p.Err != nil {
		return false
	}
	if p.Curr.Duration, p.Err = time.ParseDuration(duration + "s"); p.Err != nil {
		return false
	}
	p.Curr.Type = FileInputLine
	p.Curr.Timestep = -1
	p.Curr.Instant = time.Time{}
	p.Curr.Filename = file
	return true
}

func (p *Parser) parseCalcLine(line string) bool {
	var timeStep string
	var instant string
	var domain string
	var duration string

	groups := reCalc.FindStringSubmatch(line)
	if len(groups) == 5 {
		timeStep = groups[1]
		instant = groups[2]
		domain = groups[3]
		duration = groups[4]

	} else {
		groups = reCalcNoTimestep.FindStringSubmatch(line)
		if len(groups) == 4 {
			timeStep = "0.0"
			instant = groups[1]
			domain = groups[2]
			duration = groups[3]

		} else {
			p.Err = fmt.Errorf("malformed calculation line `%s`", line)
			return false
		}
	}

	if p.Curr.Timestep, p.Err = strconv.ParseFloat(timeStep, 64); p.Err != nil {
		return false
	}
	if p.Curr.Instant, p.Err = time.ParseInLocation("2006-01-02_15:04:05", instant, time.UTC); p.Err != nil {
		return false
	}
	if p.Curr.Domain, p.Err = strconv.ParseInt(domain, 10, 64); p.Err != nil {
		return false
	}
	if p.Curr.Duration, p.Err = time.ParseDuration(duration + "s"); p.Err != nil {
		return false
	}

	p.Curr.Type = CalcLine
	p.Curr.Filename = ""

	return true
}

type Progress struct {
	Err       error
	Val       int
	Completed bool
}

func ShowProgress(r io.Reader, start, end time.Time) chan Progress {
	ch := make(chan Progress)
	go func() {
		p := Parser{R: r}
		defer close(ch)

		duration := end.Sub(start)
		lastProgress := 0
		for p.Read() {
			if p.Curr.Type == CalcLine {
				durationSoFar := p.Curr.Instant.Sub(start)
				currProgress := int((durationSoFar * 100) / duration)
				if currProgress != lastProgress {
					ch <- Progress{Val: currProgress}
					lastProgress = currProgress
				}
			}

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
*/

func ShowDAProgress(r io.Reader, start, end time.Time) chan Progress {
	return nil
}
