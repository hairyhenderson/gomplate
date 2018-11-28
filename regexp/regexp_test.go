package regexp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFind(t *testing.T) {
	f, err := Find(`[a-z]+`, `foo bar baz`)
	assert.NoError(t, err)
	assert.Equal(t, "foo", f)

	_, err = Find(`[a-`, "")
	assert.Error(t, err)
}

func TestFindAll(t *testing.T) {
	_, err := FindAll(`[a-`, 42, "")
	assert.Error(t, err)

	testdata := []struct {
		re       string
		n        int
		in       string
		expected []string
	}{
		{`[a-z]+`, -1, `foo bar baz`, []string{"foo", "bar", "baz"}},
		{`[a-z]+`, 0, `foo bar baz`, nil},
		{`[a-z]+`, 2, `foo bar baz`, []string{"foo", "bar"}},
		{`[a-z]+`, 14, `foo bar baz`, []string{"foo", "bar", "baz"}},
	}

	for _, d := range testdata {
		f, err := FindAll(d.re, d.n, d.in)
		assert.NoError(t, err)
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
	assert.Error(t, err)

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
		assert.NoError(t, err)
		assert.Equal(t, d.expected, r)
	}
}

func TestSplit(t *testing.T) {
	_, err := Split(`[a-`, 42, "")
	assert.Error(t, err)

	testdata := []struct {
		re       string
		n        int
		in       string
		expected []string
	}{
		{`\s+`, -1, "foo  bar baz\tqux", []string{"foo", "bar", "baz", "qux"}},
		{`,`, 0, `foo bar baz`, nil},
		{` `, 2, `foo bar baz`, []string{"foo", "bar baz"}},
		{`[\s,.]`, 14, `foo bar.baz,qux`, []string{"foo", "bar", "baz", "qux"}},
	}

	for _, d := range testdata {
		f, err := Split(d.re, d.n, d.in)
		assert.NoError(t, err)
		assert.EqualValues(t, d.expected, f)
	}
}
