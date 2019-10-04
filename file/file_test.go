package file

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	tfs "gotest.tools/v3/fs"
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

	_, err = Read("/tmp/bar")
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

func TestWrite(t *testing.T) {
	oldwd, _ := os.Getwd()
	defer os.Chdir(oldwd)

	rootDir := tfs.NewDir(t, "gomplate-test")
	defer rootDir.Remove()

	newwd := rootDir.Join("the", "path", "we", "want")
	badwd := rootDir.Join("some", "other", "dir")
	fs.MkdirAll(newwd, 0755)
	fs.MkdirAll(badwd, 0755)
	newwd, _ = filepath.EvalSymlinks(newwd)
	badwd, _ = filepath.EvalSymlinks(badwd)

	err := os.Chdir(newwd)
	assert.NoError(t, err)

	err = Write("/foo", []byte("Hello world"))
	assert.Error(t, err)

	rel, err := filepath.Rel(newwd, badwd)
	assert.NoError(t, err)
	err = Write(rel, []byte("Hello world"))
	assert.Error(t, err)

	foopath := filepath.Join(newwd, "foo")
	err = Write(foopath, []byte("Hello world"))
	assert.NoError(t, err)

	out, err := ioutil.ReadFile(foopath)
	assert.NoError(t, err)
	assert.Equal(t, "Hello world", string(out))

	err = Write(foopath, []byte("truncate"))
	assert.NoError(t, err)

	out, err = ioutil.ReadFile(foopath)
	assert.NoError(t, err)
	assert.Equal(t, "truncate", string(out))

	foopath = filepath.Join(newwd, "nonexistant", "subdir", "foo")
	err = Write(foopath, []byte("Hello subdirranean world!"))
	assert.NoError(t, err)

	out, err = ioutil.ReadFile(foopath)
	assert.NoError(t, err)
	assert.Equal(t, "Hello subdirranean world!", string(out))
}

func TestAssertPathInWD(t *testing.T) {
	oldwd, _ := os.Getwd()
	defer os.Chdir(oldwd)

	err := assertPathInWD("/tmp")
	assert.Error(t, err)

	err = assertPathInWD(filepath.Join(oldwd, "subpath"))
	assert.NoError(t, err)

	err = assertPathInWD("subpath")
	assert.NoError(t, err)

	err = assertPathInWD("./subpath")
	assert.NoError(t, err)

	err = assertPathInWD(filepath.Join("..", "bogus"))
	assert.Error(t, err)

	err = assertPathInWD(filepath.Join("..", "..", "bogus"))
	assert.Error(t, err)

	base := filepath.Base(oldwd)
	err = assertPathInWD(filepath.Join("..", base))
	assert.NoError(t, err)
}
