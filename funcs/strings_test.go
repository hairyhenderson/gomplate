package funcs

import (
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateStringFuncs(t *testing.T) {
	for i := 0; i < 10; i++ {
		// Run this a bunch to catch race conditions
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			fmap := CreateStringFuncs(ctx)
			actual := fmap["strings"].(func() interface{})

			assert.Same(t, ctx, actual().(*StringFuncs).ctx)
		})
	}
}

func TestReplaceAll(t *testing.T) {
	sf := &StringFuncs{}

	assert.Equal(t, "Replaced",
		sf.ReplaceAll("Orig", "Replaced", "Orig"))
	assert.Equal(t, "ReplacedReplaced",
		sf.ReplaceAll("Orig", "Replaced", "OrigOrig"))
}

func TestIndent(t *testing.T) {
	sf := &StringFuncs{}

	testdata := []struct {
		out  string
		args []interface{}
	}{
		{" foo\n bar\n baz", []interface{}{"foo\nbar\nbaz"}},
		{"  foo\n  bar\n  baz", []interface{}{"  ", "foo\nbar\nbaz"}},
		{"---foo\n---bar\n---baz", []interface{}{3, "-", "foo\nbar\nbaz"}},
		{"   foo\n   bar\n   baz", []interface{}{3, "foo\nbar\nbaz"}},
	}

	for _, d := range testdata {
		out, err := sf.Indent(d.args...)
		assert.NoError(t, err)
		assert.Equal(t, d.out, out)
	}
}

func TestTrimPrefix(t *testing.T) {
	sf := &StringFuncs{}

	assert.Equal(t, "Bar",
		sf.TrimPrefix("Foo", "FooBar"))
}

func TestTitle(t *testing.T) {
	sf := &StringFuncs{}
	testdata := []struct {
		in  interface{}
		out string
	}{
		{``, ``},
		{`foo`, `Foo`},
		{`foo bar`, `Foo Bar`},
		{`ǉoo ǆar`, `ǈoo ǅar`},
		{`foo bar᳇baz`, `Foo Bar᳇Baz`}, // ᳇ should be treated as punctuation
		{`foo,bar&baz`, `Foo,Bar&Baz`},
	}

	for _, d := range testdata {
		up := sf.Title(d.in)
		assert.Equal(t, d.out, up)
	}
}

func TestTrunc(t *testing.T) {
	sf := &StringFuncs{}
	assert.Equal(t, "", sf.Trunc(5, ""))
	assert.Equal(t, "", sf.Trunc(0, nil))
	assert.Equal(t, "123", sf.Trunc(3, 123456789))
	assert.Equal(t, "hello, world", sf.Trunc(-1, "hello, world"))
}

func TestAbbrev(t *testing.T) {
	sf := &StringFuncs{}
	_, err := sf.Abbrev()
	assert.Error(t, err)

	_, err = sf.Abbrev("foo")
	assert.Error(t, err)

	s, err := sf.Abbrev(3, "foo")
	assert.NoError(t, err)
	assert.Equal(t, "foo", s)

	s, err = sf.Abbrev(2, 6, "foobar")
	assert.NoError(t, err)
	assert.Equal(t, "foobar", s)

	s, err = sf.Abbrev(6, 9, "foobarbazquxquux")
	assert.NoError(t, err)
	assert.Equal(t, "...baz...", s)
}

func TestSlug(t *testing.T) {
	sf := &StringFuncs{}
	s := sf.Slug(nil)
	assert.Equal(t, "nil", s)

	s = sf.Slug(0)
	assert.Equal(t, "0", s)

	s = sf.Slug(1.85e-5)
	assert.Equal(t, "1-85e-05", s)

	s = sf.Slug("Hello, World!")
	assert.Equal(t, "hello-world", s)

	s = sf.Slug("foo@example.com")
	assert.Equal(t, "fooatexample-com", s)

	s = sf.Slug("rock & roll!")
	assert.Equal(t, "rock-and-roll", s)

	s = sf.Slug("foo@example.com")
	assert.Equal(t, "fooatexample-com", s)

	s = sf.Slug(`100%`)
	assert.Equal(t, "100", s)
}

func TestSort(t *testing.T) {
	sf := &StringFuncs{ctx: context.Background()}
	in := []string{"foo", "bar", "baz"}
	out := []string{"bar", "baz", "foo"}
	assert.Equal(t, out, must(sf.Sort(in)))

	assert.Equal(t, out, must(sf.Sort([]interface{}{"foo", "bar", "baz"})))
}

func TestQuote(t *testing.T) {
	sf := &StringFuncs{}
	testdata := []struct {
		in  interface{}
		out string
	}{
		{``, `""`},
		{`foo`, `"foo"`},
		{nil, `"nil"`},
		{123.4, `"123.4"`},
		{`hello "world"`, `"hello \"world\""`},
		{`it's its`, `"it's its"`},
	}

	for _, d := range testdata {
		assert.Equal(t, d.out, sf.Quote(d.in))
	}
}

func TestShellQuote(t *testing.T) {
	sf := &StringFuncs{}
	testdata := []struct {
		in  interface{}
		out string
	}{
		// conventional cases are covered in gompstrings.ShellQuote() tests
		// we cover only cases that require type conversion or array/slice combining here
		{nil, `'nil'`},
		{123.4, `'123.4'`},
		// array and slice cases
		{[]string{}, ``},
		{[]string{"", ""}, `'' ''`},
		{[...]string{"one'two", "three four"}, `'one'"'"'two' 'three four'`},
		{[]string{"one'two", "three four"}, `'one'"'"'two' 'three four'`},
	}

	for _, d := range testdata {
		assert.Equal(t, d.out, sf.ShellQuote(d.in))
	}
}

func TestSquote(t *testing.T) {
	sf := &StringFuncs{}
	testdata := []struct {
		in  interface{}
		out string
	}{
		{``, `''`},
		{`foo`, `'foo'`},
		{nil, `'nil'`},
		{123.4, `'123.4'`},
		{`hello "world"`, `'hello "world"'`},
		{`it's its`, `'it''s its'`},
	}

	for _, d := range testdata {
		assert.Equal(t, d.out, sf.Squote(d.in))
	}
}

func TestRuneCount(t *testing.T) {
	sf := &StringFuncs{}

	n, err := sf.RuneCount("")
	assert.NoError(t, err)
	assert.Equal(t, 0, n)

	n, err = sf.RuneCount("foo")
	assert.NoError(t, err)
	assert.Equal(t, 3, n)

	n, err = sf.RuneCount("foo", "bar")
	assert.NoError(t, err)
	assert.Equal(t, 6, n)

	n, err = sf.RuneCount(42, true)
	assert.NoError(t, err)
	assert.Equal(t, 6, n)

	n, err = sf.RuneCount("😂\U0001F602")
	assert.NoError(t, err)
	assert.Equal(t, 2, n)

	n, err = sf.RuneCount("\U0001F600", 3.14)
	assert.NoError(t, err)
	assert.Equal(t, 5, n)
}
