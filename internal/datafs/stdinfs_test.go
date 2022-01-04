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
)

func TestStdinFS_Open(t *testing.T) {
	fsys, err := NewStdinFS(nil)
	assert.NoError(t, err)
	assert.IsType(t, &stdinFS{}, fsys)

	f, err := fsys.Open("foo")
	assert.NoError(t, err)
	assert.IsType(t, &stdinFile{}, f)
}

func TestStdinFile_Read(t *testing.T) {
	content := `hello world`

	f := &stdinFile{name: "foo", body: bytes.NewBufferString(content)}
	b := make([]byte, len(content))
	n, err := f.Read(b)
	assert.NoError(t, err)
	assert.Equal(t, len(content), n)
	assert.Equal(t, content, string(b))
}

func TestStdinFile_Stat(t *testing.T) {
	content := []byte(`hello world`)

	f := &stdinFile{name: "hello", body: bytes.NewReader(content)}

	fi, err := f.Stat()
	assert.NoError(t, err)
	assert.Equal(t, int64(len(content)), fi.Size())

	f = &stdinFile{name: "hello", body: &errorReader{err: fs.ErrPermission}}

	_, err = f.Stat()
	assert.ErrorIs(t, err, fs.ErrPermission)
}

func TestStdinFS(t *testing.T) {
	u, _ := url.Parse("stdin:")

	content := []byte("\nhello file\n")

	ctx := ContextWithStdin(context.Background(), bytes.NewReader(content))

	fsys, err := NewStdinFS(u)
	assert.NoError(t, err)
	assert.IsType(t, &stdinFS{}, fsys)

	_, ok := fsys.(*stdinFS)
	assert.True(t, ok)

	fsys = fsimpl.WithContextFS(ctx, fsys)

	b, err := fs.ReadFile(fsys, "foo")
	assert.NoError(t, err)
	assert.Equal(t, "\nhello file\n", string(b))

	ctx = ContextWithStdin(context.Background(), bytes.NewReader(content))
	fsys = fsimpl.WithContextFS(ctx, fsys)

	_, err = fsys.Open("..")
	assert.ErrorIs(t, err, fs.ErrInvalid)

	_, err = fs.ReadFile(fsys, "/foo")
	assert.ErrorIs(t, err, fs.ErrInvalid)

	f, err := fsys.Open("doesn't matter what it's named.txt")
	assert.NoError(t, err)

	fi, err := f.Stat()
	assert.NoError(t, err)
	assert.Equal(t, int64(len(content)), fi.Size())

	b, err = io.ReadAll(f)
	assert.NoError(t, err)
	assert.Equal(t, content, b)

	err = f.Close()
	assert.NoError(t, err)

	err = f.Close()
	assert.ErrorIs(t, err, fs.ErrClosed)

	p := make([]byte, 5)
	_, err = f.Read(p)
	assert.Error(t, err)
	assert.ErrorIs(t, err, io.EOF)
}

type errorReader struct {
	err error
}

func (r *errorReader) Read(_ []byte) (int, error) {
	return 0, r.err
}
