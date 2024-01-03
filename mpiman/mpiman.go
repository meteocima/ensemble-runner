package mpiman

import (
	"fmt"
	"strconv"
)

func ParseHosts(hosts string) (SlurmHosts, error) {
	var p parser

	if len(hosts) == 0 {
		return p.fail("empty hosts list")
	}
	p.err.Src = hosts

	for pos, c := range hosts {
		p.err.Pos = pos
		if p.parseChar(c) {
			return nil, p.err
		}
	}

	if len(p.currHost) > 0 {
		p.resHosts = append(p.resHosts, string(p.currHost))
	}

	return p.resHosts, nil
}

type SlurmHosts []string
type ParseError struct {
	Pos int
	Src string
	Msg string
}

func (e ParseError) Error() string {
	return e.Msg
}

var _ error = ParseError{}

type parser struct {
	resHosts   []string
	currHost   []rune
	currPrefix []rune
	rangeStart []rune
	err        ParseError
}

func (p *parser) parseEndRange() bool {

	if len(p.currHost) == 0 {
		p.fail("range end cannot be empty")
		return true
	}
	zeroPadding := 0
	if len(p.rangeStart) > 0 && p.rangeStart[0] == '0' {
		zeroPadding = len(p.rangeStart)
	}
	start, err := strconv.Atoi(string(p.rangeStart))
	if err != nil {
		p.fail("range start is not a number")
		return true
	}
	end, err := strconv.Atoi(string(p.currHost))
	if err != nil {
		p.fail("range end is not a number")
		return true
	}
	for i := start; i <= end; i++ {
		host := fmt.Sprintf("%s%0*d", string(p.currPrefix), zeroPadding, i)
		p.resHosts = append(p.resHosts, host)
	}
	p.rangeStart = nil
	p.currHost = nil
	return false
}

func (p *parser) fail(msg string) (SlurmHosts, error) {
	p.err.Msg = msg
	return nil, p.err
}

func (p *parser) parseChar(c rune) bool {
	switch c {
	case ',':
		return p.parseComma()
	case '[':
		return p.parseStartGroup()
	case ']':
		return p.parseEndGroup()
	case '-':
		return p.parseStartRange()
	default:
		p.currHost = append(p.currHost, c)
		return false
	}
}

func (p *parser) parseStartRange() bool {
	if len(p.currHost) == 0 {
		p.fail("range start cannot be empty")
		return true
	}
	p.rangeStart = p.currHost
	p.currHost = nil
	return false
}

func (p *parser) parseStartGroup() bool {
	p.currPrefix = p.currHost
	p.currHost = nil
	return false
}

func (p *parser) parseEndGroup() bool {
	if p.rangeStart != nil {
		if p.parseEndRange() {
			return true
		}
		p.currPrefix = nil
		return false
	}
	if len(p.currHost) == 0 && len(p.currPrefix) == 0 {
		p.fail("empty group")
		return true
	}
	host := string(p.currPrefix) + string(p.currHost)
	p.resHosts = append(p.resHosts, host)
	p.currHost = nil
	p.currPrefix = nil
	return false

}

func (p *parser) parseComma() bool {
	if p.rangeStart != nil {
		return p.parseEndRange()
	}
	if len(p.currHost) == 0 && len(p.currPrefix) == 0 {
		return false
	}
	host := string(p.currPrefix) + string(p.currHost)
	p.resHosts = append(p.resHosts, host)
	p.currHost = nil
	return false
}
