package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestShortcut(t *testing.T) {
	pc := &parseContext{}
	args := []string{"a", "b"}
	ok, nargs := shortcut.match(args, pc)
	require.True(t, ok, "shortcut always matches")
	require.Equal(t, args, nargs, "shortcut doesn't touch the passed args")
}

func TestOptsEnd(t *testing.T) {
	pc := &parseContext{}
	args := []string{"a", "b"}
	ok, nargs := optsEnd.match(args, pc)
	require.True(t, ok, "optsEnd always matches")
	require.Equal(t, args, nargs, "optsEnd doesn't touch the passed args")
	require.True(t, pc.rejectOptions, "optsEnd sets the rejectOptions flag")
}

func TestArgMatcher(t *testing.T) {
	arg := &arg{name: "X"}

	{
		pc := newParseContext()
		args := []string{"a", "b"}
		ok, nargs := arg.match(args, &pc)
		require.True(t, ok, "arg should match")
		require.Equal(t, []string{"b"}, nargs, "arg should consume the matched value")
		require.Equal(t, []string{"a"}, pc.args[arg], "arg should stored the matched value")
	}
	{
		pc := newParseContext()
		ok, _ := arg.match([]string{"-v"}, &pc)
		require.False(t, ok, "arg should not match options")
	}
	{
		pc := newParseContext()
		pc.rejectOptions = true
		ok, _ := arg.match([]string{"-v"}, &pc)
		require.True(t, ok, "arg should match options when the reject flag is set")
	}
}

func TestBoolOptMatcher(t *testing.T) {
	forceOpt := &opt{names: []string{"-f", "--force"}, value: newBoolValue(new(bool), false)}
	optMatcher := &optMatcher{
		theOne: forceOpt,
		optionsIdx: map[string]*opt{
			"-f":      forceOpt,
			"--force": forceOpt,
			"-g":      {names: []string{"-g"}, value: newBoolValue(new(bool), false)},
			"-x":      {names: []string{"-x"}, value: newBoolValue(new(bool), false)},
			"-y":      {names: []string{"-y"}, value: newBoolValue(new(bool), false)},
		},
	}
	cases := []struct {
		args  []string
		nargs []string
		val   []string
	}{
		{[]string{"-f", "x"}, []string{"x"}, []string{"true"}},
		{[]string{"-f=true", "x"}, []string{"x"}, []string{"true"}},
		{[]string{"-f=false", "x"}, []string{"x"}, []string{"false"}},
		{[]string{"--force", "x"}, []string{"x"}, []string{"true"}},
		{[]string{"--force=true", "x"}, []string{"x"}, []string{"true"}},
		{[]string{"--force=false", "x"}, []string{"x"}, []string{"false"}},
		{[]string{"-fgxy", "x"}, []string{"-gxy", "x"}, []string{"true"}},
		{[]string{"-gfxy", "x"}, []string{"-gxy", "x"}, []string{"true"}},
		{[]string{"-gxfy", "x"}, []string{"-gxy", "x"}, []string{"true"}},
		{[]string{"-gxyf", "x"}, []string{"-gxy", "x"}, []string{"true"}},
	}
	for _, cas := range cases {
		t.Logf("Testing case: %#v", cas)
		pc := newParseContext()
		ok, nargs := optMatcher.match(cas.args, &pc)
		require.True(t, ok, "opt should match")
		require.Equal(t, cas.nargs, nargs, "opt should consume the option name")
		require.Equal(t, cas.val, pc.opts[forceOpt], "true should stored as the option's value")

		pc = newParseContext()
		pc.rejectOptions = true
		nok, _ := optMatcher.match(cas.args, &pc)
		require.False(t, nok, "opt shouldn't match when rejectOptions flag is set")
	}
}

func TestOptMatcher(t *testing.T) {
	names := []string{"-f", "--force"}
	opts := []*opt{
		{names: names, value: newStringValue(new(string), "")},
		{names: names, value: newIntValue(new(int), 0)},
		{names: names, value: newStringsValue(new([]string), nil)},
		{names: names, value: newIntsValue(new([]int), nil)},
	}

	cases := []struct {
		args  []string
		nargs []string
		val   []string
	}{
		{[]string{"-f", "x"}, []string{}, []string{"x"}},
		{[]string{"-f=x", "y"}, []string{"y"}, []string{"x"}},
		{[]string{"-fx", "y"}, []string{"y"}, []string{"x"}},
		{[]string{"-afx", "y"}, []string{"-a", "y"}, []string{"x"}},
		{[]string{"-af", "x", "y"}, []string{"-a", "y"}, []string{"x"}},
		{[]string{"--force", "x"}, []string{}, []string{"x"}},
		{[]string{"--force=x", "y"}, []string{"y"}, []string{"x"}},
	}

	for _, cas := range cases {
		for _, forceOpt := range opts {
			t.Logf("Testing case: %#v with opt: %#v", cas, forceOpt)
			optMatcher := &optMatcher{
				theOne: forceOpt,
				optionsIdx: map[string]*opt{
					"-f":      forceOpt,
					"--force": forceOpt,
					"-a":      {names: []string{"-a"}, value: newBoolValue(new(bool), false)},
				},
			}

			pc := newParseContext()
			ok, nargs := optMatcher.match(cas.args, &pc)
			require.True(t, ok, "opt %#v should match args %v, %v", forceOpt, cas.args, forceOpt.isBool())
			require.Equal(t, cas.nargs, nargs, "opt should consume the option name")
			require.Equal(t, cas.val, pc.opts[forceOpt], "true should stored as the option's value")

			pc = newParseContext()
			pc.rejectOptions = true
			nok, _ := optMatcher.match(cas.args, &pc)
			require.False(t, nok, "opt shouldn't match when rejectOptions flag is set")
		}
	}
}

func TestOptsMatcher(t *testing.T) {
	opts := optsMatcher{
		options: []*opt{
			{names: []string{"-f", "--force"}, value: newBoolValue(new(bool), false)},
			{names: []string{"-g", "--green"}, value: newStringValue(new(string), "")},
		},
		optionsIndex: map[string]*opt{},
	}

	for _, o := range opts.options {
		for _, n := range o.names {
			opts.optionsIndex[n] = o
		}
	}

	cases := []struct {
		args  []string
		nargs []string
		val   [][]string
	}{
		{[]string{"-f", "x"}, []string{"x"}, [][]string{{"true"}, nil}},
		{[]string{"-f=false", "y"}, []string{"y"}, [][]string{{"false"}, nil}},
		{[]string{"--force", "x"}, []string{"x"}, [][]string{{"true"}, nil}},
		{[]string{"--force=false", "y"}, []string{"y"}, [][]string{{"false"}, nil}},

		{[]string{"-g", "x"}, []string{}, [][]string{nil, {"x"}}},
		{[]string{"-g=x", "y"}, []string{"y"}, [][]string{nil, {"x"}}},
		{[]string{"-gx", "y"}, []string{"y"}, [][]string{nil, {"x"}}},
		{[]string{"--green", "x"}, []string{}, [][]string{nil, {"x"}}},
		{[]string{"--green=x", "y"}, []string{"y"}, [][]string{nil, {"x"}}},

		{[]string{"-f", "-g", "x", "y"}, []string{"y"}, [][]string{{"true"}, {"x"}}},
		{[]string{"-g", "x", "-f", "y"}, []string{"y"}, [][]string{{"true"}, {"x"}}},
		{[]string{"-fg", "x", "y"}, []string{"y"}, [][]string{{"true"}, {"x"}}},
		{[]string{"-fgxxx", "y"}, []string{"y"}, [][]string{{"true"}, {"xxx"}}},
	}

	for _, cas := range cases {
		t.Logf("testing with args %v", cas.args)
		pc := newParseContext()
		ok, nargs := opts.match(cas.args, &pc)
		require.True(t, ok, "opts should match")
		require.Equal(t, cas.nargs, nargs, "opts should consume the option name")
		for i, opt := range opts.options {
			require.Equal(t, cas.val[i], pc.opts[opt], "the option value for %v should be stored", opt)
		}

		pc = newParseContext()
		pc.rejectOptions = true
		nok, _ := opts.match(cas.args, &pc)
		require.False(t, nok, "opts shouldn't match when rejectOptions flag is set")
	}
}
