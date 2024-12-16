package datafs

import (
	"bytes"
	"context"
	"io"
	"io/fs"
	"net/url"
	"testing"

	"github.com/hairyhenderson/go-fsimpl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStdinFS_Open(t *testing.T) {
	fsys, err := NewStdinFS(nil)
	require.NoError(t, err)
	assert.IsType(t, &stdinFS{}, fsys)

	f, err := fsys.Open("foo")
	require.NoError(t, err)
	assert.IsType(t, &stdinFile{}, f)
}

func TestStdinFile_Read(t *testing.T) {
	content := `hello world`

	f := &stdinFile{name: "foo", body: bytes.NewBufferString(content)}
	b := make([]byte, len(content))
	n, err := f.Read(b)
	require.NoError(t, err)
	assert.Equal(t, len(content), n)
	assert.Equal(t, content, string(b))
}

func TestStdinFile_Stat(t *testing.T) {
	content := []byte(`hello world`)

	f := &stdinFile{name: "hello", body: bytes.NewReader(content)}

	fi, err := f.Stat()
	require.NoError(t, err)
	assert.Equal(t, int64(len(content)), fi.Size())

	f = &stdinFile{name: "hello", body: &errorReader{err: fs.ErrPermission}}

	_, err = f.Stat()
	require.ErrorIs(t, err, fs.ErrPermission)
}

func TestStdinFS(t *testing.T) {
	u, _ := url.Parse("stdin:")

	content := []byte("\nhello file\n")

	ctx := ContextWithStdin(context.Background(), bytes.NewReader(content))

	fsys, err := NewStdinFS(u)
	require.NoError(t, err)
	assert.IsType(t, &stdinFS{}, fsys)

	_, ok := fsys.(*stdinFS)
	assert.True(t, ok)

	fsys = fsimpl.WithContextFS(ctx, fsys)

	b, err := fs.ReadFile(fsys, "foo")
	require.NoError(t, err)
	assert.Equal(t, "\nhello file\n", string(b))

	ctx = ContextWithStdin(context.Background(), bytes.NewReader(content))
	fsys = fsimpl.WithContextFS(ctx, fsys)

	_, err = fsys.Open("..")
	require.ErrorIs(t, err, fs.ErrInvalid)

	_, err = fs.ReadFile(fsys, "/foo")
	require.ErrorIs(t, err, fs.ErrInvalid)

	f, err := fsys.Open("doesn't matter what it's named.txt")
	require.NoError(t, err)

	fi, err := f.Stat()
	require.NoError(t, err)
	assert.Equal(t, int64(len(content)), fi.Size())

	b, err = io.ReadAll(f)
	require.NoError(t, err)
	assert.Equal(t, content, b)

	err = f.Close()
	require.NoError(t, err)

	err = f.Close()
	require.ErrorIs(t, err, fs.ErrClosed)

	p := make([]byte, 5)
	_, err = f.Read(p)
	require.Error(t, err)
	require.ErrorIs(t, err, io.EOF)

	t.Run("open/read multiple times", func(t *testing.T) {
		ctx := ContextWithStdin(context.Background(), bytes.NewReader(content))
		fsys = fsimpl.WithContextFS(ctx, fsys)

		for i := 0; i < 3; i++ {
			f, err := fsys.Open("foo")
			require.NoError(t, err)

			b, err := io.ReadAll(f)
			require.NoError(t, err)
			require.Equal(t, content, b, "read %d failed", i)
		}
	})

	t.Run("readFile multiple times", func(t *testing.T) {
		ctx := ContextWithStdin(context.Background(), bytes.NewReader(content))
		fsys = fsimpl.WithContextFS(ctx, fsys)

		for i := 0; i < 3; i++ {
			b, err := fs.ReadFile(fsys, "foo")
			require.NoError(t, err)
			require.Equal(t, content, b, "read %d failed", i)
		}
	})

	t.Run("open errors", func(t *testing.T) {
		ctx := ContextWithStdin(context.Background(), &errorReader{err: fs.ErrPermission})

		fsys, err := NewStdinFS(u)
		require.NoError(t, err)
		assert.IsType(t, &stdinFS{}, fsys)

		fsys = fsimpl.WithContextFS(ctx, fsys)

		_, err = fsys.Open("foo")
		require.ErrorIs(t, err, fs.ErrPermission)
	})

	t.Run("readFile errors", func(t *testing.T) {
		ctx := ContextWithStdin(context.Background(), &errorReader{err: fs.ErrPermission})

		fsys, err := NewStdinFS(u)
		require.NoError(t, err)
		assert.IsType(t, &stdinFS{}, fsys)

		fsys = fsimpl.WithContextFS(ctx, fsys)

		_, err = fs.ReadFile(fsys, "foo")
		require.ErrorIs(t, err, fs.ErrPermission)
	})
}

type errorReader struct {
	err error
}

func (r *errorReader) Read(_ []byte) (int, error) {
	return 0, r.err
}
