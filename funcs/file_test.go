package funcs

import (
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestFileExists(t *testing.T) {
	fs := afero.NewMemMapFs()
	ff := &FileFuncs{fs}

	_ = fs.Mkdir("/tmp", 0777)
	f, _ := fs.Create("/tmp/foo")
	_, _ = f.Write([]byte("foo"))

	assert.True(t, ff.Exists("/tmp/foo"))
	assert.False(t, ff.Exists("/tmp/bar"))
}

func TestFileIsDir(t *testing.T) {
	fs := afero.NewMemMapFs()
	ff := &FileFuncs{fs}

	_ = fs.Mkdir("/tmp", 0777)
	f, _ := fs.Create("/tmp/foo")
	_, _ = f.Write([]byte("foo"))

	assert.True(t, ff.IsDir("/tmp"))
	assert.False(t, ff.IsDir("/tmp/foo"))
}

func TestFileWalk(t *testing.T) {
	fs := afero.NewMemMapFs()
	ff := &FileFuncs{fs}

	_ = fs.Mkdir("/tmp", 0777)
	_ = fs.Mkdir("/tmp/bar", 0777)
	_ = fs.Mkdir("/tmp/bar/baz", 0777)
	f, _ := fs.Create("/tmp/bar/baz/foo")
	_, _ = f.Write([]byte("foo"))

	expectedLists := [][]string{{"tmp", "bar" }, {"tmp", "bar", "baz"}, {"tmp", "bar", "baz", "foo"}}
	expectedPaths := make([]string, 0)
	for _, path := range expectedLists {
		expectedPaths = append(expectedPaths, string(filepath.Separator) + filepath.Join(path...))
	}

	actualPaths, err := ff.Walk("/tmp/bar")

	assert.NoError(t, err)
	assert.Equal(t, expectedPaths, actualPaths)
}
