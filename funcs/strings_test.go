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

func TestTrimPrefix(t *testing.T) {
	sf := &StringFuncs{}

	assert.Equal(t, "Bar",
		sf.TrimPrefix("Foo", "FooBar"))
}

func TestIndent(t *testing.T) {
	sf := &StringFuncs{}
	assert.Equal(t, " foo\n bar\n baz", sf.Indent("foo\nbar\nbaz"))
	assert.Equal(t, "  foo\n  bar\n  baz", sf.Indent("  ", "foo\nbar\nbaz"))
	assert.Equal(t, "---foo\n---bar\n---baz", sf.Indent(3, "-", "foo\nbar\nbaz"))
	assert.Equal(t, "   foo\n   bar\n   baz", sf.Indent(3, "foo\nbar\nbaz"))
}
