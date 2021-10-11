package funcs

import (
	"context"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestCreateFileFuncs(t *testing.T) {
	t.Parallel()

	for i := 0; i < 10; i++ {
		// Run this a bunch to catch race conditions
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			fmap := CreateFileFuncs(ctx)
			actual := fmap["file"].(func() interface{})

			assert.Same(t, ctx, actual().(*FileFuncs).ctx)
		})
	}
}

func TestFileExists(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	ff := &FileFuncs{fs: fs}

	_ = fs.Mkdir("/tmp", 0777)
	f, _ := fs.Create("/tmp/foo")
	_, _ = f.Write([]byte("foo"))

	assert.True(t, ff.Exists("/tmp/foo"))
	assert.False(t, ff.Exists("/tmp/bar"))
}

func TestFileIsDir(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	ff := &FileFuncs{fs: fs}

	_ = fs.Mkdir("/tmp", 0777)
	f, _ := fs.Create("/tmp/foo")
	_, _ = f.Write([]byte("foo"))

	assert.True(t, ff.IsDir("/tmp"))
	assert.False(t, ff.IsDir("/tmp/foo"))
}

func TestFileWalk(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	ff := &FileFuncs{fs: fs}

	_ = fs.Mkdir("/tmp", 0777)
	_ = fs.Mkdir("/tmp/bar", 0777)
	_ = fs.Mkdir("/tmp/bar/baz", 0777)
	f, _ := fs.Create("/tmp/bar/baz/foo")
	_, _ = f.Write([]byte("foo"))

	expectedLists := [][]string{{"tmp"}, {"tmp", "bar"}, {"tmp", "bar", "baz"}, {"tmp", "bar", "baz", "foo"}}
	expectedPaths := make([]string, 0)
	for _, path := range expectedLists {
		expectedPaths = append(expectedPaths, string(filepath.Separator)+filepath.Join(path...))
	}

	actualPaths, err := ff.Walk(string(filepath.Separator) + "tmp")

	assert.NoError(t, err)
	assert.Equal(t, expectedPaths, actualPaths)
}
