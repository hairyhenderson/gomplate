package cli

import (
	"fmt"
	"strings"
)

type step struct {
	do      func()
	success *step
	error   *step
	desc    string
}

func (s *step) run(p interface{}) {
	s.callDo(p)

	switch {
	case s.success != nil:
		s.success.run(p)
	case p == nil:
		return
	default:
		if code, ok := p.(exit); ok {
			exiter(int(code))
			return
		}
		panic(p)
	}
}

func (s *step) callDo(p interface{}) {
	if s.do == nil {
		return
	}
	defer func() {
		if e := recover(); e != nil {
			if s.error == nil {
				panic(p)
			}
			s.error.run(e)
		}
	}()
	s.do()
}

func (s *step) dot() string {
	trs := flowDot(s, map[*step]bool{})
	return fmt.Sprintf("digraph G {\n\trankdir=LR\n%s\n}\n", strings.Join(trs, "\n"))
}

func flowDot(s *step, visited map[*step]bool) []string {
	res := []string{}
	if visited[s] {
		return res
	}
	visited[s] = true

	if s.success != nil {
		res = append(res, fmt.Sprintf("\t\"%s\" -> \"%s\" [label=\"ok\"]", s.desc, s.success.desc))
		res = append(res, flowDot(s.success, visited)...)
	}
	if s.error != nil {
		res = append(res, fmt.Sprintf("\t\"%s\" -> \"%s\" [label=\"ko\"]", s.desc, s.error.desc))
		res = append(res, flowDot(s.error, visited)...)
	}
	return res
}
