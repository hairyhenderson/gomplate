package cli

import (
	"sort"
	"strings"

	"fmt"
)

type state struct {
	id          int
	terminal    bool
	transitions transitions
	cmd         *Cmd
}

type transition struct {
	matcher upMatcher
	next    *state
}

type transitions []*transition

func (t transitions) Len() int      { return len(t) }
func (t transitions) Swap(i, j int) { t[i], t[j] = t[j], t[i] }
func (t transitions) Less(i, j int) bool {
	a, _ := t[i].matcher, t[j].matcher
	switch a.(type) {
	case upShortcut:
		return false
	case upOptsEnd:
		return false
	case *arg:
		return false
	default:
		return true
	}

}

var _id = 0

func newState(cmd *Cmd) *state {
	_id++
	return &state{_id, false, []*transition{}, cmd}
}

func (s *state) t(matcher upMatcher, next *state) *state {
	s.transitions = append(s.transitions, &transition{matcher, next})
	return next
}

func (s *state) has(tr *transition) bool {
	for _, t := range s.transitions {
		if t.next == tr.next && t.matcher == tr.matcher {
			return true
		}
	}
	return false
}

func incoming(s, into *state, visited map[*state]bool) []*transition {
	res := []*transition{}
	if visited[s] {
		return res
	}
	visited[s] = true

	for _, tr := range s.transitions {
		if tr.next == into {
			res = append(res, tr)
		}
		res = append(res, incoming(tr.next, into, visited)...)
	}
	return res
}

func removeTransitionAt(idx int, arr transitions) transitions {
	res := make([]*transition, len(arr)-1)
	copy(res, arr[:idx])
	copy(res[idx:], arr[idx+1:])
	return res
}

func (s *state) simplify() {
	simplify(s, s, map[*state]bool{})
}

func simplify(start, s *state, visited map[*state]bool) {
	if visited[s] {
		return
	}
	visited[s] = true
	for _, tr := range s.transitions {
		simplify(start, tr.next, visited)
	}
	for s.simplifySelf(start) {
	}
}

func (s *state) simplifySelf(start *state) bool {
	for idx, tr := range s.transitions {
		if _, ok := tr.matcher.(upShortcut); ok {
			next := tr.next
			s.transitions = removeTransitionAt(idx, s.transitions)
			for _, tr := range next.transitions {
				if !s.has(tr) {
					s.transitions = append(s.transitions, tr)
				}
			}
			if next.terminal {
				s.terminal = true
			}
			return true
		}
	}
	return false
}

func (s *state) dot() string {
	trs := dot(s, map[*state]bool{})
	return fmt.Sprintf("digraph G {\n\trankdir=LR\n%s\n}\n", strings.Join(trs, "\n"))
}

func dot(s *state, visited map[*state]bool) []string {
	res := []string{}
	if visited[s] {
		return res
	}
	visited[s] = true

	for _, tr := range s.transitions {
		res = append(res, fmt.Sprintf("\tS%d -> S%d [label=\"%v\"]", s.id, tr.next.id, tr.matcher))
		res = append(res, dot(tr.next, visited)...)
	}
	if s.terminal {
		res = append(res, fmt.Sprintf("\tS%d [peripheries=2]", s.id))
	}
	return res
}

type parseContext struct {
	args          map[*arg][]string
	opts          map[*opt][]string
	rejectOptions bool
}

func newParseContext() parseContext {
	return parseContext{map[*arg][]string{}, map[*opt][]string{}, false}
}

func (pc parseContext) merge(o parseContext) {
	for k, vs := range o.args {
		pc.args[k] = append(pc.args[k], vs...)
	}

	for k, vs := range o.opts {
		pc.opts[k] = append(pc.opts[k], vs...)
	}
}

func (s *state) parse(args []string) error {
	pc := newParseContext()
	ok, err := s.apply(args, pc)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("incorrect usage")
	}

	for opt, vs := range pc.opts {
		multiValued, ok := opt.value.(multiValued)
		if ok && opt.valueSetFromEnv {
			multiValued.Clear()
			opt.valueSetFromEnv = false
		}
		for _, v := range vs {
			if err := opt.value.Set(v); err != nil {
				return err
			}
		}

		if opt.valueSetByUser != nil {
			*opt.valueSetByUser = true
		}
	}

	for arg, vs := range pc.args {
		multiValued, ok := arg.value.(multiValued)
		if ok && arg.valueSetFromEnv {
			multiValued.Clear()
			arg.valueSetFromEnv = false
		}
		for _, v := range vs {
			if err := arg.value.Set(v); err != nil {
				return err
			}
		}

		if arg.valueSetByUser != nil {
			*arg.valueSetByUser = true
		}
	}

	return nil
}

func (s *state) apply(args []string, pc parseContext) (bool, error) {
	if s.terminal && len(args) == 0 {
		return true, nil
	}
	sort.Sort(s.transitions)

	if len(args) > 0 {
		arg := args[0]

		if !pc.rejectOptions && arg == "--" {
			pc.rejectOptions = true
			args = args[1:]
		}
	}

	type match struct {
		tr  *transition
		rem []string
		pc  parseContext
	}

	matches := []*match{}
	for _, tr := range s.transitions {
		fresh := newParseContext()
		fresh.rejectOptions = pc.rejectOptions
		if ok, rem := tr.matcher.match(args, &fresh); ok {
			matches = append(matches, &match{tr, rem, fresh})
		}
	}

	for _, m := range matches {
		ok, err := m.tr.next.apply(m.rem, m.pc)
		if err != nil {
			return false, err
		}
		if ok {
			pc.merge(m.pc)
			return true, nil
		}
	}
	return false, nil
}
