//go:build !windows
// +build !windows

package funcs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilePathFuncs(t *testing.T) {
	t.Parallel()

	f := &FilePathFuncs{}
	assert.Equal(t, "bar", f.Base("foo/bar"))
	assert.Equal(t, "bar", f.Base("/foo/bar"))

	assert.Equal(t, "/foo/baz", f.Clean("/foo/bar/../baz"))
	assert.Equal(t, "foo", f.Dir("foo/bar"))
	assert.Equal(t, ".txt", f.Ext("/foo/bar/baz.txt"))
	assert.False(t, f.IsAbs("foo/bar"))
	assert.True(t, f.IsAbs("/foo/bar"))
	assert.Equal(t, "foo/bar/qux", f.Join("foo", "bar", "baz", "..", "qux"))
	m, _ := f.Match("*.txt", "foo.json")
	assert.False(t, m)
	m, _ = f.Match("*.txt", "foo.txt")
	assert.True(t, m)
	r, _ := f.Rel("/foo/bar", "/foo/bar/baz")
	assert.Equal(t, "baz", r)
	assert.Equal(t, []string{"/foo/bar/", "baz"}, f.Split("/foo/bar/baz"))
	assert.Equal(t, "", f.VolumeName("/foo/bar"))
}
