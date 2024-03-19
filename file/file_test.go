package file

import (
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/hairyhenderson/gomplate/v4/internal/datafs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRead(t *testing.T) {
	oldFS := fsys
	defer func() { fsys = oldFS }()
	fsys = datafs.WrapWdFS(fstest.MapFS{
		"tmp":     &fstest.MapFile{Mode: fs.ModeDir | 0o777},
		"tmp/foo": &fstest.MapFile{Data: []byte("foo")},
	})

	actual, err := Read("/tmp/foo")
	require.NoError(t, err)
	assert.Equal(t, "foo", actual)

	_, err = Read("/tmp/bar")
	require.Error(t, err)
}

func TestReadDir(t *testing.T) {
	oldFS := fsys
	defer func() { fsys = oldFS }()
	fsys = datafs.WrapWdFS(fstest.MapFS{
		"tmp":          &fstest.MapFile{Mode: fs.ModeDir | 0o777},
		"tmp/foo":      &fstest.MapFile{Data: []byte("foo")},
		"tmp/bar":      &fstest.MapFile{Data: []byte("bar")},
		"tmp/baz":      &fstest.MapFile{Data: []byte("baz")},
		"tmp/qux":      &fstest.MapFile{Mode: fs.ModeDir | 0o777},
		"tmp/qux/quux": &fstest.MapFile{Data: []byte("quux")},
	})

	actual, err := ReadDir("/tmp")
	require.NoError(t, err)
	assert.Equal(t, []string{"bar", "baz", "foo", "qux"}, actual)

	_, err = ReadDir("/tmp/foo")
	require.Error(t, err)
}
