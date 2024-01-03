package mpiman

import (
	"fmt"
	"strconv"
)

type SlurmHosts []string
type ParseError struct {
	pos int
	src string
	msg string
}

func (e ParseError) Error() string {
	return e.msg
}

var _ error = ParseError{}

func ParseHosts(hosts string) (SlurmHosts, error) {
	var resHosts []string
	var currHost []rune
	var currPrefix []rune
	var rangeStart []rune
	err := ParseError{
		pos: 0,
		src: hosts,
	}

	parseRange := func(c rune) bool {
		fail := func(msg string) bool {
			err.msg = msg
			return true
		}
		if len(currHost) == 0 {
			return fail("range end cannot be empty")
		}
		zeroPadding := 0
		if len(rangeStart) > 0 && rangeStart[0] == '0' {
			zeroPadding = len(rangeStart)
		}
		start, err := strconv.Atoi(string(rangeStart))
		if err != nil {
			return fail("range start is not a number")
		}
		end, err := strconv.Atoi(string(currHost))
		if err != nil {
			return fail("range end is not a number")
		}
		for i := start; i <= end; i++ {
			host := fmt.Sprintf("%s%0*d", string(currPrefix), zeroPadding, i)
			resHosts = append(resHosts, host)
		}
		rangeStart = nil
		currHost = nil
		return false
	}
	fail := func(msg string) (SlurmHosts, error) {
		err.msg = msg
		return nil, err
	}
	if len(hosts) == 0 {
		return fail("empty hosts list")
	}
	for pos, c := range hosts {
		err.pos = pos
		switch c {
		case ',':
			if rangeStart != nil {
				if parseRange(c) {
					return nil, err
				}
				continue
			}
			if len(currHost) == 0 && len(currPrefix) == 0 {
				continue
			}
			host := string(currPrefix) + string(currHost)
			resHosts = append(resHosts, host)
			currHost = nil
		case '[':
			currPrefix = currHost
			currHost = nil
		case ']':
			if rangeStart != nil {
				if parseRange(c) {
					return nil, err
				}
				currPrefix = nil
				continue
			}
			if len(currHost) == 0 && len(currPrefix) == 0 {
				return fail("empty group")
			}
			host := string(currPrefix) + string(currHost)
			resHosts = append(resHosts, host)
			currHost = nil
			currPrefix = nil
		case '-':
			if len(currHost) == 0 {
				return fail("range start cannot be empty")
			}
			rangeStart = currHost
			currHost = nil
		default:
			currHost = append(currHost, c)
		}
	}
	if len(currHost) > 0 {
		resHosts = append(resHosts, string(currHost))
	}

	return resHosts, nil
}
