package cli

import (
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStringOpt(t *testing.T) {
	cmd := &Cmd{optionsIdx: map[string]*opt{}}
	a := cmd.String(StringOpt{Name: "a", Value: "test", Desc: ""})
	require.Equal(t, "test", *a)

	os.Setenv("B", "")
	b := cmd.String(StringOpt{Name: "b", Value: "test", EnvVar: "B", Desc: ""})
	require.Equal(t, "test", *b)

	os.Setenv("B", "mow")
	b = cmd.String(StringOpt{Name: "b", Value: "test", EnvVar: "B", Desc: ""})
	require.Equal(t, "mow", *b)

	os.Setenv("B", "")
	os.Setenv("C", "cli")
	os.Setenv("D", "mow")
	b = cmd.String(StringOpt{Name: "b", Value: "test", EnvVar: "B C D", Desc: ""})
	require.Equal(t, "cli", *b)
}

func TestBoolOpt(t *testing.T) {
	cmd := &Cmd{optionsIdx: map[string]*opt{}}
	a := cmd.Bool(BoolOpt{Name: "a", Value: true, Desc: ""})
	require.True(t, *a)

	os.Setenv("B", "")
	b := cmd.Bool(BoolOpt{Name: "b", Value: false, EnvVar: "B", Desc: ""})
	require.False(t, *b)

	trueValues := []string{"1", "true", "TRUE"}
	for _, tv := range trueValues {
		os.Setenv("B", tv)
		b = cmd.Bool(BoolOpt{Name: "b", Value: false, EnvVar: "B", Desc: ""})
		require.True(t, *b, "env=%s", tv)
	}

	falseValues := []string{"0", "false", "FALSE", "xyz"}
	for _, tv := range falseValues {
		os.Setenv("B", tv)
		b = cmd.Bool(BoolOpt{Name: "b", Value: false, EnvVar: "B", Desc: ""})
		require.False(t, *b, "env=%s", tv)
	}

	os.Setenv("B", "")
	os.Setenv("C", "false")
	os.Setenv("D", "true")
	b = cmd.Bool(BoolOpt{Name: "b", Value: true, EnvVar: "B C D", Desc: ""})
	require.False(t, *b)
}

func TestIntOpt(t *testing.T) {
	cmd := &Cmd{optionsIdx: map[string]*opt{}}
	a := cmd.Int(IntOpt{Name: "a", Value: -1, Desc: ""})
	require.Equal(t, -1, *a)

	os.Setenv("B", "")
	b := cmd.Int(IntOpt{Name: "b", Value: -1, EnvVar: "B", Desc: ""})
	require.Equal(t, -1, *b)

	goodValues := []int{1, 0, 33}
	for _, tv := range goodValues {
		os.Setenv("B", strconv.Itoa(tv))
		b := cmd.Int(IntOpt{Name: "b", Value: -1, EnvVar: "B", Desc: ""})
		require.Equal(t, tv, *b, "env=%s", tv)
	}

	badValues := []string{"", "b", "q1", "_"}
	for _, tv := range badValues {
		os.Setenv("B", tv)
		b := cmd.Int(IntOpt{Name: "b", Value: -1, EnvVar: "B", Desc: ""})
		require.Equal(t, -1, *b, "env=%s", tv)
	}

	os.Setenv("B", "")
	os.Setenv("C", "42")
	os.Setenv("D", "666")
	b = cmd.Int(IntOpt{Name: "b", Value: -1, EnvVar: "B C D", Desc: ""})
	require.Equal(t, 42, *b)
}

func TestStringsOpt(t *testing.T) {
	cmd := &Cmd{optionsIdx: map[string]*opt{}}
	v := []string{"test"}
	a := cmd.Strings(StringsOpt{Name: "a", Value: v, Desc: ""})
	require.Equal(t, v, *a)

	os.Setenv("B", "")
	b := cmd.Strings(StringsOpt{Name: "b", Value: v, EnvVar: "B", Desc: ""})
	require.Equal(t, v, *b)

	os.Setenv("B", "mow")
	b = cmd.Strings(StringsOpt{Name: "b", Value: nil, EnvVar: "B", Desc: ""})
	require.Equal(t, []string{"mow"}, *b)

	os.Setenv("B", "mow, cli")
	b = cmd.Strings(StringsOpt{Name: "b", Value: nil, EnvVar: "B", Desc: ""})
	require.Equal(t, []string{"mow", "cli"}, *b)

	os.Setenv("B", "")
	os.Setenv("C", "test")
	os.Setenv("D", "xxx")
	b = cmd.Strings(StringsOpt{Name: "b", Value: nil, EnvVar: "B C D", Desc: ""})
	require.Equal(t, v, *b)
}

func TestIntsOpt(t *testing.T) {
	cmd := &Cmd{optionsIdx: map[string]*opt{}}
	vi := []int{42}
	a := cmd.Ints(IntsOpt{Name: "a", Value: vi, Desc: ""})
	require.Equal(t, vi, *a)

	os.Setenv("B", "")
	b := cmd.Ints(IntsOpt{Name: "b", Value: vi, EnvVar: "B", Desc: ""})
	require.Equal(t, vi, *b)

	os.Setenv("B", "666")
	b = cmd.Ints(IntsOpt{Name: "b", Value: nil, EnvVar: "B", Desc: ""})
	require.Equal(t, []int{666}, *b)

	os.Setenv("B", "1, 2 , 3")
	b = cmd.Ints(IntsOpt{Name: "b", Value: nil, EnvVar: "B", Desc: ""})
	require.Equal(t, []int{1, 2, 3}, *b)

	os.Setenv("B", "")
	os.Setenv("C", "abc")
	os.Setenv("D", "1, abc")
	os.Setenv("E", "42")
	os.Setenv("F", "666")
	b = cmd.Ints(IntsOpt{Name: "b", Value: nil, EnvVar: "B C D E F", Desc: ""})
	require.Equal(t, vi, *b)
}
