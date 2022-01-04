package datafs

import (
	"io/fs"
	"net/url"
	"os"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
)

func TestEnvFS_Open(t *testing.T) {
	fsys, err := NewEnvFS(nil)
	assert.NoError(t, err)
	assert.IsType(t, &envFS{}, fsys)

	f, err := fsys.Open("foo")
	assert.NoError(t, err)
	assert.IsType(t, &envFile{}, f)
}

func TestEnvFile_Read(t *testing.T) {
	content := `hello world`
	os.Setenv("HELLO_WORLD", "hello world")
	defer os.Unsetenv("HELLO_WORLD")

	f := &envFile{name: "HELLO_WORLD"}
	b := make([]byte, len(content))
	n, err := f.Read(b)
	assert.NoError(t, err)
	assert.Equal(t, len(content), n)
	assert.Equal(t, content, string(b))

	fsys := fstest.MapFS{}
	fsys["foo/bar/baz.txt"] = &fstest.MapFile{Data: []byte("\nhello world\n")}

	os.Setenv("FOO_FILE", "/foo/bar/baz.txt")
	defer os.Unsetenv("FOO_FILE")

	f = &envFile{name: "FOO", locfs: fsys}

	b = make([]byte, len(content))
	t.Logf("b len is %d", len(b))
	n, err = f.Read(b)
	t.Logf("b len is %d", len(b))
	assert.NoError(t, err)
	assert.Equal(t, len(content), n)
	assert.Equal(t, content, string(b))
}

func TestEnvFile_Stat(t *testing.T) {
	content := []byte(`hello world`)
	os.Setenv("HELLO_WORLD", "hello world")
	defer os.Unsetenv("HELLO_WORLD")

	f := &envFile{name: "HELLO_WORLD"}

	fi, err := f.Stat()
	assert.NoError(t, err)
	assert.Equal(t, int64(len(content)), fi.Size())

	fsys := fstest.MapFS{}
	fsys["foo/bar/baz.txt"] = &fstest.MapFile{Data: []byte("\nhello world\n")}

	os.Setenv("FOO_FILE", "/foo/bar/baz.txt")
	defer os.Unsetenv("FOO_FILE")

	f = &envFile{name: "FOO", locfs: fsys}

	fi, err = f.Stat()
	assert.NoError(t, err)
	assert.Equal(t, int64(len(content)), fi.Size())
}

func TestEnvFS(t *testing.T) {
	u, _ := url.Parse("env:")

	lfsys := fstest.MapFS{}
	lfsys["foo/bar/baz.txt"] = &fstest.MapFile{Data: []byte("\nhello file\n")}

	fsys, err := NewEnvFS(u)
	assert.NoError(t, err)
	assert.IsType(t, &envFS{}, fsys)

	envfs, ok := fsys.(*envFS)
	assert.True(t, ok)
	envfs.locfs = lfsys

	os.Setenv("FOO_FILE", "/foo/bar/baz.txt")
	defer os.Unsetenv("FOO_FILE")

	b, err := fs.ReadFile(fsys, "FOO")
	assert.NoError(t, err)
	assert.Equal(t, "hello file", string(b))

	os.Setenv("FOO", "hello world")
	defer os.Unsetenv("FOO")

	b, err = fs.ReadFile(fsys, "FOO")
	assert.NoError(t, err)
	assert.Equal(t, "hello world", string(b))

	assert.NoError(t, fstest.TestFS(fsys, "FOO", "FOO_FILE", "HOME"))
}
