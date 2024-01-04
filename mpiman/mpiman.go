package mpiman

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/exp/maps"
)

func ParseSlurmNodes(hosts string) (SlurmNodes, error) {
	var p parser
	p.resHosts = NewSlurmNodes()
	if len(hosts) == 0 {
		p.fail("empty hosts list")
		return SlurmNodes{}, p.err
	}
	p.err.Src = hosts

	for pos, c := range hosts {
		p.err.Pos = pos
		if p.parseChar(c) {
			return SlurmNodes{}, p.err
		}
	}

	if len(p.currHost) > 0 {
		p.resHosts.Nodes[string(p.currHost)] = true
	}

	return p.resHosts, nil
}
func NewSlurmNodes() SlurmNodes {
	return SlurmNodes{
		Nodes: make(map[string]bool),
		Lock:  &sync.Mutex{},
	}
}

type SlurmNodes struct {
	Nodes map[string]bool
	Lock  *sync.Mutex
}

type SlurmNodesList []string

func (sn SlurmNodes) All() SlurmNodesList {
	res := make(SlurmNodesList, 0, len(sn.Nodes))
	for k := range sn.Nodes {
		res = append(res, k)
	}
	sort.Strings(res)
	return res
}

func (lst SlurmNodesList) String() string {
	return fmt.Sprintf("-hosts %s", strings.Join(lst, ","))
}

func (sn SlurmNodes) Dispose(lst SlurmNodesList) {
	sn.Lock.Lock()
	defer sn.Lock.Unlock()

	for _, k := range lst {
		sn.Nodes[k] = true
	}
}

func (sn SlurmNodes) FindFreeNodes(n int) (SlurmNodesList, bool) {
	res := make(SlurmNodesList, 0, len(sn.Nodes))
	sn.Lock.Lock()
	defer sn.Lock.Unlock()
	nodes := maps.Keys(sn.Nodes)
	sort.Strings(nodes)

	for _, node := range nodes {
		free := sn.Nodes[node]
		if free {
			res = append(res, node)
			sn.Nodes[node] = false
		}
		if len(res) == n {
			break
		}
	}
	if len(res) < n {
		return nil, false
	}
	sort.Strings(res)
	return res, true
}

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
	resHosts   SlurmNodes
	currHost   []rune
	currPrefix []rune
	rangeStart []rune
	err        ParseError
}

func (p *parser) parseEndRange() bool {

	if len(p.currHost) == 0 {
		return p.fail("range end cannot be empty")
	}
	zeroPadding := 0
	if len(p.rangeStart) > 0 && p.rangeStart[0] == '0' {
		zeroPadding = len(p.rangeStart)
	}
	start, err := strconv.Atoi(string(p.rangeStart))
	if err != nil {
		return p.fail("range start is not a number")
	}
	end, err := strconv.Atoi(string(p.currHost))
	if err != nil {
		return p.fail("range end is not a number")
	}
	for i := start; i <= end; i++ {
		host := fmt.Sprintf("%s%0*d", string(p.currPrefix), zeroPadding, i)
		p.resHosts.Nodes[host] = true
	}
	p.rangeStart = nil
	p.currHost = nil
	return false
}

func (p *parser) fail(msg string) bool {
	p.err.Msg = msg
	return true
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
		return p.fail("range start cannot be empty")
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
		return p.fail("empty group")
	}
	host := string(p.currPrefix) + string(p.currHost)
	p.resHosts.Nodes[host] = true
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
	p.resHosts.Nodes[host] = true
	p.currHost = nil
	return false
}
