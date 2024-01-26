//go:build windows
// +build windows

package funcs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilePathFuncs(t *testing.T) {
	t.Parallel()

	f := &FilePathFuncs{}
	assert.Equal(t, "bar", f.Base(`foo\bar`))
	assert.Equal(t, "bar", f.Base("C:/foo/bar"))
	assert.Equal(t, "bar", f.Base(`C:\foo\bar`))

	assert.Equal(t, `C:\foo\baz`, f.Clean(`C:\foo\bar\..\baz`))
	assert.Equal(t, "foo", f.Dir(`foo\bar`))
	assert.Equal(t, ".txt", f.Ext(`C:\foo\bar\baz.txt`))
	assert.False(t, f.IsAbs(`foo\bar`))
	assert.True(t, f.IsAbs(`C:\foo\bar`))
	assert.False(t, f.IsAbs(`\foo\bar`))
	assert.Equal(t, `foo\bar\qux`, f.Join("foo", "bar", "baz", "..", "qux"))
	m, _ := f.Match("*.txt", "foo.json")
	assert.False(t, m)
	m, _ = f.Match("*.txt", "foo.txt")
	assert.True(t, m)
	r, _ := f.Rel(`C:\foo\bar`, `C:\foo\bar\baz`)
	assert.Equal(t, "baz", r)
	assert.Equal(t, []string{`C:\foo\bar\`, "baz"}, f.Split(`C:\foo\bar\baz`))
	assert.Equal(t, "D:", f.VolumeName(`D:\foo\bar`))
}
