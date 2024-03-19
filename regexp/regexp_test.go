package regexp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFind(t *testing.T) {
	f, err := Find(`[a-z]+`, `foo bar baz`)
	require.NoError(t, err)
	assert.Equal(t, "foo", f)

	_, err = Find(`[a-`, "")
	require.Error(t, err)
}

func TestFindAll(t *testing.T) {
	_, err := FindAll(`[a-`, 42, "")
	require.Error(t, err)

	testdata := []struct {
		re       string
		in       string
		expected []string
		n        int
	}{
		{`[a-z]+`, `foo bar baz`, []string{"foo", "bar", "baz"}, -1},
		{`[a-z]+`, `foo bar baz`, nil, 0},
		{`[a-z]+`, `foo bar baz`, []string{"foo", "bar"}, 2},
		{`[a-z]+`, `foo bar baz`, []string{"foo", "bar", "baz"}, 14},
	}

	for _, d := range testdata {
		f, err := FindAll(d.re, d.n, d.in)
		require.NoError(t, err)
		assert.EqualValues(t, d.expected, f)
	}
}

func TestMatch(t *testing.T) {
	assert.True(t, Match(`^[a-z]+\[[0-9]+\]$`, "adam[23]"))
	assert.True(t, Match(`^[a-z]+\[[0-9]+\]$`, "eve[7]"))
	assert.False(t, Match(`^[a-z]+\[[0-9]+\]$`, "Job[48]"))
	assert.False(t, Match(`^[a-z]+\[[0-9]+\]$`, "snakey"))
}

func TestReplace(t *testing.T) {
	testdata := []struct {
		expected    string
		expression  string
		replacement string
		input       string
	}{
		{"-T-T-", "a(x*)b", "T", "-ab-axxb-"},
		{"--xx-", "a(x*)b", "$1", "-ab-axxb-"},
		{"---", "a(x*)b", "$1W", "-ab-axxb-"},
		{"-W-xxW-", "a(x*)b", "${1}W", "-ab-axxb-"},
		{"Turing, Alan", "(?P<first>[a-zA-Z]+) (?P<last>[a-zA-Z]+)", "${last}, ${first}", "Alan Turing"},
	}
	for _, d := range testdata {
		assert.Equal(t, d.expected, Replace(d.expression, d.replacement, d.input))
	}
}

func TestReplaceLiteral(t *testing.T) {
	_, err := ReplaceLiteral(`[a-`, "", "")
	require.Error(t, err)

	testdata := []struct {
		expected    string
		expression  string
		replacement string
		input       string
	}{
		{"-T-T-", "a(x*)b", "T", "-ab-axxb-"},
		{"-$1-$1-", "a(x*)b", "$1", "-ab-axxb-"},
		{"-$1W-$1W-", "a(x*)b", "$1W", "-ab-axxb-"},
		{"-${1}W-${1}W-", "a(x*)b", "${1}W", "-ab-axxb-"},
		{"${last}, ${first}", "(?P<first>[a-zA-Z]+) (?P<last>[a-zA-Z]+)", "${last}, ${first}", "Alan Turing"},
	}
	for _, d := range testdata {
		r, err := ReplaceLiteral(d.expression, d.replacement, d.input)
		require.NoError(t, err)
		assert.Equal(t, d.expected, r)
	}
}

func TestSplit(t *testing.T) {
	_, err := Split(`[a-`, 42, "")
	require.Error(t, err)

	testdata := []struct {
		re       string
		in       string
		expected []string
		n        int
	}{
		{`\s+`, "foo  bar baz\tqux", []string{"foo", "bar", "baz", "qux"}, -1},
		{`,`, `foo bar baz`, nil, 0},
		{` `, `foo bar baz`, []string{"foo", "bar baz"}, 2},
		{`[\s,.]`, `foo bar.baz,qux`, []string{"foo", "bar", "baz", "qux"}, 14},
	}

	for _, d := range testdata {
		f, err := Split(d.re, d.n, d.in)
		require.NoError(t, err)
		assert.EqualValues(t, d.expected, f)
	}
}

func TestQuoteMeta(t *testing.T) {
	assert.Equal(t, `foo\{\(\\`, QuoteMeta(`foo{(\`))
}
