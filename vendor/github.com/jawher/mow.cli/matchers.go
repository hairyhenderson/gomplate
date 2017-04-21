package cli

import (
	"fmt"
	"strings"
)

type upMatcher interface {
	match(args []string, c *parseContext) (bool, []string)
}

type upShortcut bool

func (u upShortcut) match(args []string, c *parseContext) (bool, []string) {
	return true, args
}

func (u upShortcut) String() string {
	return "*"
}

type upOptsEnd bool

func (u upOptsEnd) match(args []string, c *parseContext) (bool, []string) {
	c.rejectOptions = true
	return true, args
}

func (u upOptsEnd) String() string {
	return "--"
}

const (
	shortcut = upShortcut(true)
	optsEnd  = upOptsEnd(true)
)

func (arg *arg) match(args []string, c *parseContext) (bool, []string) {
	if len(args) == 0 {
		return false, args
	}
	if !c.rejectOptions && strings.HasPrefix(args[0], "-") && args[0] != "-" {
		return false, args
	}
	c.args[arg] = append(c.args[arg], args[0])
	return true, args[1:]
}

type optMatcher struct {
	theOne     *opt
	optionsIdx map[string]*opt
}

func (o *optMatcher) match(args []string, c *parseContext) (bool, []string) {
	if len(args) == 0 || c.rejectOptions {
		return o.theOne.valueSetFromEnv, args
	}

	idx := 0
	for idx < len(args) {
		arg := args[idx]
		switch {
		case arg == "-":
			idx++
		case arg == "--":
			return o.theOne.valueSetFromEnv, nil
		case strings.HasPrefix(arg, "--"):
			matched, consumed, nargs := o.matchLongOpt(args, idx, c)

			if matched {
				return true, nargs
			}
			if consumed == 0 {
				return o.theOne.valueSetFromEnv, args
			}
			idx += consumed

		case strings.HasPrefix(arg, "-"):
			matched, consumed, nargs := o.matchShortOpt(args, idx, c)
			if matched {
				return true, nargs
			}
			if consumed == 0 {
				return o.theOne.valueSetFromEnv, args
			}
			idx += consumed

		default:
			return o.theOne.valueSetFromEnv, args
		}
	}
	return o.theOne.valueSetFromEnv, args
}

func (o *optMatcher) matchLongOpt(args []string, idx int, c *parseContext) (bool, int, []string) {
	arg := args[idx]
	kv := strings.Split(arg, "=")
	name := kv[0]
	opt, found := o.optionsIdx[name]
	if !found {
		return false, 0, args
	}

	switch {
	case len(kv) == 2:
		if opt != o.theOne {
			return false, 1, args
		}
		value := kv[1]
		c.opts[o.theOne] = append(c.opts[o.theOne], value)
		return true, 1, removeStringAt(idx, args)
	case opt.isBool():
		if opt != o.theOne {
			return false, 1, args
		}
		c.opts[o.theOne] = append(c.opts[o.theOne], "true")
		return true, 1, removeStringAt(idx, args)
	default:
		if len(args[idx:]) < 2 {
			return false, 0, args
		}
		if opt != o.theOne {
			return false, 2, args
		}
		value := args[idx+1]
		if strings.HasPrefix(value, "-") {
			return false, 0, args
		}
		c.opts[o.theOne] = append(c.opts[o.theOne], value)
		return true, 2, removeStringsBetween(idx, idx+1, args)
	}
}

func (o *optMatcher) matchShortOpt(args []string, idx int, c *parseContext) (bool, int, []string) {
	arg := args[idx]
	if len(arg) < 2 {
		return false, 0, args
	}

	if strings.HasPrefix(arg[2:], "=") {
		name := arg[0:2]
		opt, _ := o.optionsIdx[name]
		if opt == o.theOne {
			value := arg[3:]
			if value == "" {
				return false, 0, args
			}
			c.opts[o.theOne] = append(c.opts[o.theOne], value)
			return true, 1, removeStringAt(idx, args)
		}

		return false, 1, args
	}

	rem := arg[1:]

	remIdx := 0
	for len(rem[remIdx:]) > 0 {
		name := "-" + rem[remIdx:remIdx+1]

		opt, found := o.optionsIdx[name]
		if !found {
			return false, 0, args
		}

		if opt.isBool() {
			if opt != o.theOne {
				remIdx++
				continue
			}

			c.opts[o.theOne] = append(c.opts[o.theOne], "true")
			newRem := rem[:remIdx] + rem[remIdx+1:]
			if newRem == "" {
				return true, 1, removeStringAt(idx, args)
			}
			return true, 0, replaceStringAt(idx, "-"+newRem, args)
		}

		value := rem[remIdx+1:]
		if value == "" {
			if len(args[idx+1:]) == 0 {
				return false, 0, args
			}
			if opt != o.theOne {
				return false, 2, args
			}

			value = args[idx+1]
			if strings.HasPrefix(value, "-") {
				return false, 0, args
			}
			c.opts[o.theOne] = append(c.opts[o.theOne], value)

			newRem := rem[:remIdx]
			if newRem == "" {
				return true, 2, removeStringsBetween(idx, idx+1, args)
			}

			nargs := replaceStringAt(idx, "-"+newRem, args)

			return true, 1, removeStringAt(idx+1, nargs)
		}

		if opt != o.theOne {
			return false, 1, args
		}
		c.opts[o.theOne] = append(c.opts[o.theOne], value)
		newRem := rem[:remIdx]
		if newRem == "" {
			return true, 1, removeStringAt(idx, args)
		}
		return true, 0, replaceStringAt(idx, "-"+newRem, args)

	}

	return false, 1, args
}

type optsMatcher struct {
	options      []*opt
	optionsIndex map[string]*opt
}

func (om optsMatcher) try(args []string, c *parseContext) (bool, []string) {
	if len(args) == 0 || c.rejectOptions {
		return false, args
	}
	for _, o := range om.options {
		if ok, nargs := (&optMatcher{theOne: o, optionsIdx: om.optionsIndex}).match(args, c); ok {
			return ok, nargs
		}
	}
	return false, args
}

func (om optsMatcher) match(args []string, c *parseContext) (bool, []string) {
	ok, nargs := om.try(args, c)
	if !ok {
		return false, args
	}

	for {
		ok, nnargs := om.try(nargs, c)
		if !ok {
			return true, nargs
		}
		nargs = nnargs
	}
}

func (om optsMatcher) String() string {
	return fmt.Sprintf("Opts(%v)", om.options)
}

func removeStringAt(idx int, arr []string) []string {
	res := make([]string, len(arr)-1)
	copy(res, arr[:idx])
	copy(res[idx:], arr[idx+1:])
	return res
}

func removeStringsBetween(from, to int, arr []string) []string {
	res := make([]string, len(arr)-(to-from+1))
	copy(res, arr[:from])
	copy(res[from:], arr[to+1:])
	return res
}

func replaceStringAt(idx int, with string, arr []string) []string {
	res := make([]string, len(arr))
	copy(res, arr[:idx])
	res[idx] = with
	copy(res[idx+1:], arr[idx+1:])
	return res
}
