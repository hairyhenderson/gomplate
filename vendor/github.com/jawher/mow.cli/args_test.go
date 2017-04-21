package cli

import (
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStringArg(t *testing.T) {
	cmd := &Cmd{argsIdx: map[string]*arg{}}
	a := cmd.String(StringArg{Name: "a", Value: "test"})
	require.Equal(t, "test", *a)

	os.Setenv("B", "")
	b := cmd.String(StringArg{Name: "b", Value: "test", EnvVar: "B"})
	require.Equal(t, "test", *b)

	os.Setenv("B", "mow")
	b = cmd.String(StringArg{Name: "b", Value: "test", EnvVar: "B"})
	require.Equal(t, "mow", *b)

	os.Setenv("B", "")
	os.Setenv("C", "cli")
	os.Setenv("D", "mow")
	b = cmd.String(StringArg{Name: "b", Value: "test", EnvVar: "B C D"})
	require.Equal(t, "cli", *b)
}

func TestBoolArg(t *testing.T) {
	cmd := &Cmd{argsIdx: map[string]*arg{}}
	a := cmd.Bool(BoolArg{Name: "a", Value: true, Desc: ""})
	require.True(t, *a)

	os.Setenv("B", "")
	b := cmd.Bool(BoolArg{Name: "b", Value: false, EnvVar: "B", Desc: ""})
	require.False(t, *b)

	trueValues := []string{"1", "true", "TRUE"}
	for _, tv := range trueValues {
		os.Setenv("B", tv)
		b = cmd.Bool(BoolArg{Name: "b", Value: false, EnvVar: "B", Desc: ""})
		require.True(t, *b, "env=%s", tv)
	}

	falseValues := []string{"0", "false", "FALSE", "xyz"}
	for _, tv := range falseValues {
		os.Setenv("B", tv)
		b = cmd.Bool(BoolArg{Name: "b", Value: false, EnvVar: "B", Desc: ""})
		require.False(t, *b, "env=%s", tv)
	}

	os.Setenv("B", "")
	os.Setenv("C", "false")
	os.Setenv("D", "true")
	b = cmd.Bool(BoolArg{Name: "b", Value: true, EnvVar: "B C D", Desc: ""})
	require.False(t, *b)
}

func TestIntArg(t *testing.T) {
	cmd := &Cmd{argsIdx: map[string]*arg{}}
	a := cmd.Int(IntArg{Name: "a", Value: -1, Desc: ""})
	require.Equal(t, -1, *a)

	os.Setenv("B", "")
	b := cmd.Int(IntArg{Name: "b", Value: -1, EnvVar: "B", Desc: ""})
	require.Equal(t, -1, *b)

	goodValues := []int{1, 0, 33}
	for _, tv := range goodValues {
		os.Setenv("B", strconv.Itoa(tv))
		b := cmd.Int(IntArg{Name: "b", Value: -1, EnvVar: "B", Desc: ""})
		require.Equal(t, tv, *b, "env=%s", tv)
	}

	badValues := []string{"", "b", "q1", "_"}
	for _, tv := range badValues {
		os.Setenv("B", tv)
		b := cmd.Int(IntArg{Name: "b", Value: -1, EnvVar: "B", Desc: ""})
		require.Equal(t, -1, *b, "env=%s", tv)
	}

	os.Setenv("B", "")
	os.Setenv("C", "42")
	os.Setenv("D", "666")
	b = cmd.Int(IntArg{Name: "b", Value: -1, EnvVar: "B C D", Desc: ""})
	require.Equal(t, 42, *b)
}

func TestStringsArg(t *testing.T) {
	cmd := &Cmd{argsIdx: map[string]*arg{}}
	v := []string{"test"}
	a := cmd.Strings(StringsArg{Name: "a", Value: v, Desc: ""})
	require.Equal(t, v, *a)

	os.Setenv("B", "")
	b := cmd.Strings(StringsArg{Name: "b", Value: v, EnvVar: "B", Desc: ""})
	require.Equal(t, v, *b)

	os.Setenv("B", "mow")
	b = cmd.Strings(StringsArg{Name: "b", Value: nil, EnvVar: "B", Desc: ""})
	require.Equal(t, []string{"mow"}, *b)

	os.Setenv("B", "mow, cli")
	b = cmd.Strings(StringsArg{Name: "b", Value: nil, EnvVar: "B", Desc: ""})
	require.Equal(t, []string{"mow", "cli"}, *b)

	os.Setenv("B", "")
	os.Setenv("C", "test")
	os.Setenv("D", "xxx")
	b = cmd.Strings(StringsArg{Name: "b", Value: nil, EnvVar: "B C D", Desc: ""})
	require.Equal(t, v, *b)
}

func TestIntsArg(t *testing.T) {
	cmd := &Cmd{argsIdx: map[string]*arg{}}

	vi := []int{42}
	a := cmd.Ints(IntsArg{Name: "a", Value: vi, Desc: ""})
	require.Equal(t, vi, *a)

	os.Setenv("B", "")
	b := cmd.Ints(IntsArg{Name: "b", Value: vi, EnvVar: "B", Desc: ""})
	require.Equal(t, vi, *b)

	os.Setenv("B", "666")
	b = cmd.Ints(IntsArg{Name: "b", Value: nil, EnvVar: "B", Desc: ""})
	require.Equal(t, []int{666}, *b)

	os.Setenv("B", "1, 2 , 3")
	b = cmd.Ints(IntsArg{Name: "b", Value: nil, EnvVar: "B", Desc: ""})
	require.Equal(t, []int{1, 2, 3}, *b)

	os.Setenv("B", "")
	os.Setenv("C", "abc")
	os.Setenv("D", "1, abc")
	os.Setenv("E", "42")
	os.Setenv("F", "666")
	b = cmd.Ints(IntsArg{Name: "b", Value: nil, EnvVar: "B C D E F", Desc: ""})
	require.Equal(t, vi, *b)
}
