package datafs

import (
	"io/fs"
	"net/url"
	"os"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	t.Setenv("HELLO_WORLD", "hello world")

	f := &envFile{name: "HELLO_WORLD"}
	b := make([]byte, len(content))
	n, err := f.Read(b)
	assert.NoError(t, err)
	assert.Equal(t, len(content), n)
	assert.Equal(t, content, string(b))

	fsys := fstest.MapFS{}
	fsys["foo/bar/baz.txt"] = &fstest.MapFile{Data: []byte("\nhello world\n")}

	t.Setenv("FOO_FILE", "/foo/bar/baz.txt")

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
	t.Setenv("HELLO_WORLD", "hello world")

	f := &envFile{name: "HELLO_WORLD"}

	fi, err := f.Stat()
	assert.NoError(t, err)
	assert.Equal(t, int64(len(content)), fi.Size())

	fsys := fstest.MapFS{}
	fsys["foo/bar/baz.txt"] = &fstest.MapFile{Data: []byte("\nhello world\n")}

	t.Setenv("FOO_FILE", "/foo/bar/baz.txt")

	f = &envFile{name: "FOO", locfs: fsys}

	fi, err = f.Stat()
	assert.NoError(t, err)
	assert.Equal(t, int64(len(content)), fi.Size())
}

func TestEnvFS(t *testing.T) {
	t.Cleanup(func() { environ = os.Environ })

	u, _ := url.Parse("env:")

	lfsys := fstest.MapFS{}
	lfsys["foo/bar/baz.txt"] = &fstest.MapFile{Data: []byte("\nhello file\n")}

	fsys, err := NewEnvFS(u)
	assert.NoError(t, err)
	assert.IsType(t, &envFS{}, fsys)

	envfs, ok := fsys.(*envFS)
	assert.True(t, ok)
	envfs.locfs = lfsys

	t.Setenv("FOO_FILE", "/foo/bar/baz.txt")

	b, err := fs.ReadFile(fsys, "FOO")
	assert.NoError(t, err)
	assert.Equal(t, "hello file", string(b))

	t.Setenv("FOO", "hello world")

	b, err = fs.ReadFile(fsys, "FOO")
	assert.NoError(t, err)
	assert.Equal(t, "hello world", string(b))

	assert.NoError(t, fstest.TestFS(fsys, "FOO", "FOO_FILE"))
}

func TestEnvFile_ReadDir(t *testing.T) {
	t.Cleanup(func() { environ = os.Environ })

	t.Run("name must be .", func(t *testing.T) {
		f := &envFile{name: "foo"}
		_, err := f.ReadDir(-1)
		require.Error(t, err)
	})

	t.Run("empty env should return empty dir", func(t *testing.T) {
		f := &envFile{name: "."}
		environ = func() []string { return []string{} }
		des, err := f.ReadDir(-1)
		require.NoError(t, err)
		assert.Empty(t, des)
	})

	t.Run("non-empty env should return dir with entries", func(t *testing.T) {
		f := &envFile{name: "."}
		environ = func() []string { return []string{"FOO=bar", "BAR=quux"} }
		des, err := f.ReadDir(-1)
		require.NoError(t, err)
		require.Len(t, des, 2)
		assert.Equal(t, "FOO", des[0].Name())
		assert.Equal(t, "BAR", des[1].Name())
	})

	t.Run("deal with odd Windows env vars like '=C:=C:\tmp'", func(t *testing.T) {
		f := &envFile{name: "."}
		environ = func() []string { return []string{"FOO=bar", "=C:=C:\\tmp", "BAR=quux"} }
		des, err := f.ReadDir(-1)
		require.NoError(t, err)
		require.Len(t, des, 2)
		assert.Equal(t, "FOO", des[0].Name())
		assert.Equal(t, "BAR", des[1].Name())
	})
}
