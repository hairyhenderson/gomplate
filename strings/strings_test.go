package strings

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIndent(t *testing.T) {
	actual := "hello\nworld\n!"
	expected := "  hello\n  world\n  !"
	assert.Equal(t, actual, Indent(0, "  ", actual))
	assert.Equal(t, expected, Indent(1, "  ", actual))
	assert.Equal(t, "\n", Indent(1, "  ", "\n"))
	assert.Equal(t, "  foo\n", Indent(1, "  ", "foo\n"))
	assert.Equal(t, "   foo", Indent(1, "   ", "foo"))
	assert.Equal(t, "   foo", Indent(3, " ", "foo"))
}

func TestTrunc(t *testing.T) {
	assert.Equal(t, "", Trunc(5, ""))
	assert.Equal(t, "", Trunc(0, "hello, world"))
	assert.Equal(t, "hello", Trunc(5, "hello, world"))
	assert.Equal(t, "hello, world", Trunc(12, "hello, world"))
	assert.Equal(t, "hello, world", Trunc(42, "hello, world"))
	assert.Equal(t, "hello, world", Trunc(-1, "hello, world"))
}

func TestSort(t *testing.T) {
	in := []string{}
	expected := []string{}
	assert.EqualValues(t, expected, Sort(in))

	in = []string{"c", "a", "b"}
	expected = []string{"a", "b", "c"}
	assert.EqualValues(t, expected, Sort(in))

	in = []string{"42", "45", "18"}
	expected = []string{"18", "42", "45"}
	assert.EqualValues(t, expected, Sort(in))
}
