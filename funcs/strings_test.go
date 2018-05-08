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
	assert.Equal(t, " foo\n bar\n baz", sf.Indent("foo\nbar\nbaz"))
	assert.Equal(t, "  foo\n  bar\n  baz", sf.Indent("  ", "foo\nbar\nbaz"))
	assert.Equal(t, "---foo\n---bar\n---baz", sf.Indent(3, "-", "foo\nbar\nbaz"))
	assert.Equal(t, "   foo\n   bar\n   baz", sf.Indent(3, "foo\nbar\nbaz"))
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
