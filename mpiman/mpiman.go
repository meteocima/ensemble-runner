package mpiman

import (
	"fmt"
	"log"
	"strconv"
)

type SlurmHosts []string

func ParseSlurmHosts(hosts string) SlurmHosts {
	var resHosts []string
	var currHost []rune
	var currPrefix []rune
	var rangeStart []rune
	var c rune
	parseRange := func() {
		if len(currHost) == 0 {
			log.Panicf("Invalid host string: %s", hosts)
		}
		zeroPadding := 0
		if len(rangeStart) > 0 && rangeStart[0] == '0' {
			zeroPadding = len(rangeStart)
		}
		start, err := strconv.Atoi(string(rangeStart))
		if err != nil {
			log.Panicf("Invalid host string: %s", hosts)
		}
		end, err := strconv.Atoi(string(currHost))
		if err != nil {
			log.Panicf("Invalid host string: %s", hosts)
		}
		for i := start; i <= end; i++ {
			host := fmt.Sprintf("%s%0*d", string(currPrefix), zeroPadding, i)
			resHosts = append(resHosts, host)
		}
		rangeStart = nil
		currHost = nil
	}
	for _, c = range hosts {
		switch c {
		case ',':
			if rangeStart != nil {
				parseRange()
				continue
			}
			if len(currHost) == 0 && len(currPrefix) == 0 {
				log.Panicf("Invalid host string: %s", hosts)
			}
			host := string(currPrefix) + string(currHost)
			resHosts = append(resHosts, host)
			currHost = nil
		case '[':
			currPrefix = currHost
			currHost = nil
		case ']':
			if rangeStart != nil {
				parseRange()
				currPrefix = nil
				continue
			}
			if len(currHost) == 0 && len(currPrefix) == 0 {
				log.Panicf("Invalid host string: %s", hosts)
			}
			host := string(currPrefix) + string(currHost)
			resHosts = append(resHosts, host)
			currHost = nil
			currPrefix = nil
		case '-':
			if len(currHost) == 0 {
				log.Panicf("Invalid host string: %s", hosts)
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
	return resHosts
}
