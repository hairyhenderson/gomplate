// this is in a separate package so WriteFile can be more thoroughly tested
// without involving an import cycle with datafs
package iohelpers_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/hack-pad/hackpadfs"
	osfs "github.com/hack-pad/hackpadfs/os"
	"github.com/hairyhenderson/gomplate/v4/internal/datafs"
	"github.com/hairyhenderson/gomplate/v4/internal/iohelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tfs "gotest.tools/v3/fs"
)

func TestWrite(t *testing.T) {
	oldwd, _ := os.Getwd()
	defer os.Chdir(oldwd)

	rootDir := tfs.NewDir(t, "gomplate-test")
	t.Cleanup(rootDir.Remove)

	// we want to use a real filesystem here, so we can test interactions with
	// the current working directory
	fsys := datafs.WrapWdFS(osfs.NewFS())

	newwd := rootDir.Join("the", "path", "we", "want")
	badwd := rootDir.Join("some", "other", "dir")
	hackpadfs.MkdirAll(fsys, newwd, 0o755)
	hackpadfs.MkdirAll(fsys, badwd, 0o755)
	newwd, _ = filepath.EvalSymlinks(newwd)
	badwd, _ = filepath.EvalSymlinks(badwd)

	err := os.Chdir(newwd)
	require.NoError(t, err)

	err = iohelpers.WriteFile(fsys, "/foo", []byte("Hello world"))
	require.Error(t, err)

	rel, err := filepath.Rel(newwd, badwd)
	require.NoError(t, err)
	err = iohelpers.WriteFile(fsys, rel, []byte("Hello world"))
	require.Error(t, err)

	foopath := filepath.Join(newwd, "foo")
	err = iohelpers.WriteFile(fsys, foopath, []byte("Hello world"))
	require.NoError(t, err)

	out, err := fs.ReadFile(fsys, foopath)
	require.NoError(t, err)
	assert.Equal(t, "Hello world", string(out))

	err = iohelpers.WriteFile(fsys, foopath, []byte("truncate"))
	require.NoError(t, err)

	out, err = fs.ReadFile(fsys, foopath)
	require.NoError(t, err)
	assert.Equal(t, "truncate", string(out))

	foopath = filepath.Join(newwd, "nonexistent", "subdir", "foo")
	err = iohelpers.WriteFile(fsys, foopath, []byte("Hello subdirranean world!"))
	require.NoError(t, err)

	out, err = fs.ReadFile(fsys, foopath)
	require.NoError(t, err)
	assert.Equal(t, "Hello subdirranean world!", string(out))
}
