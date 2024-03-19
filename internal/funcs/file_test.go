package funcs

import (
	"bytes"
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"testing/fstest"

	"github.com/hack-pad/hackpadfs"
	osfs "github.com/hack-pad/hackpadfs/os"
	"github.com/hairyhenderson/gomplate/v4/internal/datafs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tfs "gotest.tools/v3/fs"
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

			assert.Equal(t, ctx, actual().(*FileFuncs).ctx)
		})
	}
}

func TestFileExists(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{
		"tmp":     &fstest.MapFile{Mode: fs.ModeDir | 0o777},
		"tmp/foo": &fstest.MapFile{Data: []byte("foo")},
	}
	ff := &FileFuncs{fs: datafs.WrapWdFS(fsys)}

	assert.True(t, ff.Exists("/tmp/foo"))
	assert.False(t, ff.Exists("/tmp/bar"))
}

func TestFileIsDir(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{
		"tmp":     &fstest.MapFile{Mode: fs.ModeDir | 0o777},
		"tmp/foo": &fstest.MapFile{Data: []byte("foo")},
	}

	ff := &FileFuncs{fs: datafs.WrapWdFS(fsys)}

	assert.True(t, ff.IsDir("/tmp"))
	assert.False(t, ff.IsDir("/tmp/foo"))
}

func TestFileWalk(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{
		"tmp":             &fstest.MapFile{Mode: fs.ModeDir | 0o777},
		"tmp/bar":         &fstest.MapFile{Mode: fs.ModeDir | 0o777},
		"tmp/bar/baz":     &fstest.MapFile{Mode: fs.ModeDir | 0o777},
		"tmp/bar/baz/foo": &fstest.MapFile{Data: []byte("foo")},
	}

	ff := &FileFuncs{fs: datafs.WrapWdFS(fsys)}

	expectedLists := [][]string{{"tmp"}, {"tmp", "bar"}, {"tmp", "bar", "baz"}, {"tmp", "bar", "baz", "foo"}}
	expectedPaths := make([]string, 0)
	for _, path := range expectedLists {
		expectedPaths = append(expectedPaths, string(filepath.Separator)+filepath.Join(path...))
	}

	actualPaths, err := ff.Walk(string(filepath.Separator) + "tmp")

	require.NoError(t, err)
	assert.Equal(t, expectedPaths, actualPaths)
}

func TestReadDir(t *testing.T) {
	fsys := fs.FS(fstest.MapFS{
		"tmp":          &fstest.MapFile{Mode: fs.ModeDir | 0o777},
		"tmp/foo":      &fstest.MapFile{Data: []byte("foo")},
		"tmp/bar":      &fstest.MapFile{Data: []byte("bar")},
		"tmp/baz":      &fstest.MapFile{Data: []byte("baz")},
		"tmp/qux":      &fstest.MapFile{Mode: fs.ModeDir | 0o777},
		"tmp/qux/quux": &fstest.MapFile{Data: []byte("quux")},
	})

	fsys = datafs.WrapWdFS(fsys)

	ff := &FileFuncs{
		ctx: context.Background(),
		fs:  fsys,
	}

	actual, err := ff.ReadDir("/tmp")
	require.NoError(t, err)
	assert.Equal(t, []string{"bar", "baz", "foo", "qux"}, actual)

	_, err = ff.ReadDir("/tmp/foo")
	require.Error(t, err)
}

func TestWrite(t *testing.T) {
	oldwd, _ := os.Getwd()
	defer os.Chdir(oldwd)

	rootDir := tfs.NewDir(t, "gomplate-test")
	t.Cleanup(rootDir.Remove)

	// we want to use a real filesystem here, so we can test interactions with
	// the current working directory
	fsys := datafs.WrapWdFS(osfs.NewFS())

	f := &FileFuncs{
		ctx: context.Background(),
		fs:  fsys,
	}

	newwd := rootDir.Join("the", "path", "we", "want")
	badwd := rootDir.Join("some", "other", "dir")
	hackpadfs.MkdirAll(fsys, newwd, 0o755)
	hackpadfs.MkdirAll(fsys, badwd, 0o755)
	newwd, _ = filepath.EvalSymlinks(newwd)
	badwd, _ = filepath.EvalSymlinks(badwd)

	err := os.Chdir(newwd)
	require.NoError(t, err)

	_, err = f.Write("/foo", []byte("Hello world"))
	require.Error(t, err)

	rel, err := filepath.Rel(newwd, badwd)
	require.NoError(t, err)
	_, err = f.Write(rel, []byte("Hello world"))
	require.Error(t, err)

	foopath := filepath.Join(newwd, "foo")
	_, err = f.Write(foopath, []byte("Hello world"))
	require.NoError(t, err)

	out, err := fs.ReadFile(fsys, foopath)
	require.NoError(t, err)
	assert.Equal(t, "Hello world", string(out))

	_, err = f.Write(foopath, []byte("truncate"))
	require.NoError(t, err)

	out, err = fs.ReadFile(fsys, foopath)
	require.NoError(t, err)
	assert.Equal(t, "truncate", string(out))

	foopath = filepath.Join(newwd, "nonexistant", "subdir", "foo")
	_, err = f.Write(foopath, "Hello subdirranean world!")
	require.NoError(t, err)

	out, err = fs.ReadFile(fsys, foopath)
	require.NoError(t, err)
	assert.Equal(t, "Hello subdirranean world!", string(out))

	_, err = f.Write(foopath, bytes.NewBufferString("Hello from a byte buffer!"))
	require.NoError(t, err)

	out, err = fs.ReadFile(fsys, foopath)
	require.NoError(t, err)
	assert.Equal(t, "Hello from a byte buffer!", string(out))
}
