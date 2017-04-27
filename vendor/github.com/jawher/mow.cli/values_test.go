package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBoolParam(t *testing.T) {
	var into bool

	param := newBoolValue(&into, false)

	require.True(t, param.IsBoolFlag())

	cases := []struct {
		input  string
		err    bool
		result bool
		string string
	}{
		{"true", false, true, "true"},
		{"false", false, false, "false"},
		{"123", true, false, ""},
		{"", true, false, ""},
	}

	for _, cas := range cases {
		t.Logf("testing with %q", cas.input)

		err := param.Set(cas.input)

		if cas.err {
			require.Error(t, err, "value %q should have returned an error", cas.input)
			continue
		}

		require.Equal(t, cas.result, into)
		require.Equal(t, cas.string, param.String())
	}
}

func TestStringParam(t *testing.T) {
	var into string

	param := newStringValue(&into, "")

	cases := []struct {
		input  string
		string string
	}{
		{"a", `"a"`},
		{"", `""`},
	}

	for _, cas := range cases {
		t.Logf("testing with %q", cas.input)

		err := param.Set(cas.input)

		require.NoError(t, err)

		require.Equal(t, cas.input, into)
		require.Equal(t, cas.string, param.String())
	}
}

func TestIntParam(t *testing.T) {
	var into int

	param := newIntValue(&into, 0)

	cases := []struct {
		input  string
		err    bool
		result int
		string string
	}{
		{"12", false, 12, "12"},
		{"0", false, 0, "0"},
		{"01", false, 1, "1"},
		{"", true, 0, ""},
		{"abc", true, 0, ""},
	}

	for _, cas := range cases {
		t.Logf("testing with %q", cas.input)

		err := param.Set(cas.input)

		if cas.err {
			require.Error(t, err, "value %q should have returned an error", cas.input)
			continue
		}

		require.Equal(t, cas.result, into)
		require.Equal(t, cas.string, param.String())
	}
}

func TestStringsParam(t *testing.T) {
	into := []string{}
	param := newStringsValue(&into, nil)

	param.Set("a")
	param.Set("b")

	require.Equal(t, []string{"a", "b"}, into)
	require.Equal(t, `["a", "b"]`, param.String())

	param.Clear()

	require.Empty(t, into)
}

func TestIntsParam(t *testing.T) {
	into := []int{}
	param := newIntsValue(&into, nil)

	err := param.Set("1")
	require.NoError(t, err)

	err = param.Set("2")
	require.NoError(t, err)

	require.Equal(t, []int{1, 2}, into)

	require.Equal(t, `[1, 2]`, param.String())

	err = param.Set("c")
	require.Error(t, err)
	require.Equal(t, []int{1, 2}, into)

	param.Clear()

	require.Empty(t, into)
}
