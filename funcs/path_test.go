package funcs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathFuncs(t *testing.T) {
	p := PathNS()
	assert.Equal(t, "bar", p.Base("foo/bar"))
	assert.Equal(t, "bar", p.Base("/foo/bar"))

	assert.Equal(t, "/foo/baz", p.Clean("/foo/bar/../baz"))
	assert.Equal(t, "foo", p.Dir("foo/bar"))
	assert.Equal(t, ".txt", p.Ext("/foo/bar/baz.txt"))
	assert.False(t, false, p.IsAbs("foo/bar"))
	assert.True(t, p.IsAbs("/foo/bar"))
	assert.Equal(t, "foo/bar/qux", p.Join("foo", "bar", "baz", "..", "qux"))
	m, _ := p.Match("*.txt", "foo.json")
	assert.False(t, m)
	m, _ = p.Match("*.txt", "foo.txt")
	assert.True(t, m)
	assert.Equal(t, []string{"/foo/bar/", "baz"}, p.Split("/foo/bar/baz"))
}
