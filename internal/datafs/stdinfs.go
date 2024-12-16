package datafs

import (
	"bytes"
	"context"
	"io"
	"io/fs"
	"net/url"
	"time"

	"github.com/hairyhenderson/go-fsimpl"
)

// NewStdinFS returns a filesystem (an fs.FS) that can be used to read data from
// standard input (os.Stdin).
func NewStdinFS(_ *url.URL) (fs.FS, error) {
	return &stdinFS{ctx: context.Background()}, nil
}

type stdinFS struct {
	ctx  context.Context
	data []byte
}

//nolint:gochecknoglobals
var StdinFS = fsimpl.FSProviderFunc(NewStdinFS, "stdin")

var (
	_ fs.FS         = (*stdinFS)(nil)
	_ fs.ReadFileFS = (*stdinFS)(nil)
	_ withContexter = (*stdinFS)(nil)
)

func (f stdinFS) WithContext(ctx context.Context) fs.FS {
	fsys := f
	fsys.ctx = ctx

	return &fsys
}

func (f *stdinFS) Open(name string) (fs.File, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{
			Op:   "open",
			Path: name,
			Err:  fs.ErrInvalid,
		}
	}

	if err := f.readData(); err != nil {
		return nil, &fs.PathError{
			Op:   "open",
			Path: name,
			Err:  err,
		}
	}

	return &stdinFile{name: name, body: bytes.NewReader(f.data)}, nil
}

func (f *stdinFS) ReadFile(name string) ([]byte, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{
			Op:   "readFile",
			Path: name,
			Err:  fs.ErrInvalid,
		}
	}

	if err := f.readData(); err != nil {
		return nil, &fs.PathError{
			Op:   "readFile",
			Path: name,
			Err:  err,
		}
	}

	return f.data, nil
}

func (f *stdinFS) readData() error {
	if f.data != nil {
		return nil
	}

	stdin := StdinFromContext(f.ctx)

	b, err := io.ReadAll(stdin)
	if err != nil {
		return err
	}

	f.data = b

	return nil
}

type stdinFile struct {
	body io.Reader
	name string
}

var _ fs.File = (*stdinFile)(nil)

func (f *stdinFile) Close() error {
	if f.body == nil {
		return &fs.PathError{Op: "close", Path: f.name, Err: fs.ErrClosed}
	}

	f.body = nil
	return nil
}

func (f *stdinFile) stdinReader() (int, error) {
	b, err := io.ReadAll(f.body)
	if err != nil {
		return 0, err
	}

	f.body = bytes.NewReader(b)

	return len(b), err
}

func (f *stdinFile) Stat() (fs.FileInfo, error) {
	n, err := f.stdinReader()
	if err != nil {
		return nil, err
	}

	return FileInfo(f.name, int64(n), 0o444, time.Time{}, ""), nil
}

func (f *stdinFile) Read(p []byte) (int, error) {
	if f.body == nil {
		return 0, io.EOF
	}

	return f.body.Read(p)
}
