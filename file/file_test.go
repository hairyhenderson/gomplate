package file

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestRead(t *testing.T) {
	origfs := fs
	defer func() { fs = origfs }()
	fs = afero.NewMemMapFs()
	_ = fs.Mkdir("/tmp", 0777)
	f, _ := fs.Create("/tmp/foo")
	_, _ = f.Write([]byte("foo"))

	actual, err := Read("/tmp/foo")
	assert.NoError(t, err)
	assert.Equal(t, "foo", actual)

	actual, err = Read("/tmp/bar")
	assert.Error(t, err)
}

func TestReadDir(t *testing.T) {
	origfs := fs
	defer func() { fs = origfs }()
	fs = afero.NewMemMapFs()
	fs.Mkdir("/tmp", 0777)
	fs.Create("/tmp/foo")
	fs.Create("/tmp/bar")
	fs.Create("/tmp/baz")
	fs.Mkdir("/tmp/qux", 0777)
	fs.Create("/tmp/qux/quux")

	actual, err := ReadDir("/tmp")
	assert.NoError(t, err)
	assert.Equal(t, []string{"bar", "baz", "foo", "qux"}, actual)

	_, err = ReadDir("/tmp/foo")
	assert.Error(t, err)
}
