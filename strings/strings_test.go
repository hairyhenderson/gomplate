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
