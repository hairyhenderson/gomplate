package funcs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
		args []interface{}
		out  string
	}{
		{[]interface{}{"foo\nbar\nbaz"}, " foo\n bar\n baz"},
		{[]interface{}{"  ", "foo\nbar\nbaz"}, "  foo\n  bar\n  baz"},
		{[]interface{}{3, "-", "foo\nbar\nbaz"}, "---foo\n---bar\n---baz"},
		{[]interface{}{3, "foo\nbar\nbaz"}, "   foo\n   bar\n   baz"},
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
	sf := &StringFuncs{}
	in := []string{"foo", "bar", "baz"}
	out := []string{"bar", "baz", "foo"}
	assert.Equal(t, out, must(sf.Sort(in)))

	assert.Equal(t, out, must(sf.Sort([]interface{}{"foo", "bar", "baz"})))
}
