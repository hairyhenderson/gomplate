package datafs

import (
	"errors"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/hack-pad/hackpadfs"

	"github.com/stretchr/testify/assert"
)

func TestGetenvFsys(t *testing.T) {
	fsys := fs.FS(fstest.MapFS{
		"tmp":            &fstest.MapFile{Mode: fs.ModeDir | 0o777},
		"tmp/foo":        &fstest.MapFile{Data: []byte("foo")},
		"tmp/unreadable": &fstest.MapFile{Data: []byte("foo"), Mode: 0o000},
	})
	fsys = WrapWdFS(fsys)

	t.Setenv("FOO_FILE", "/tmp/foo")
	assert.Equal(t, "foo", GetenvFsys(fsys, "FOO", "bar"))

	t.Setenv("FOO_FILE", "/tmp/missing")
	assert.Equal(t, "bar", GetenvFsys(fsys, "FOO", "bar"))

	fsys = writeOnly(fsys)
	t.Setenv("FOO_FILE", "/tmp/unreadable")
	assert.Equal(t, "bar", GetenvFsys(fsys, "FOO", "bar"))
}

func TestExpandEnvFsys(t *testing.T) {
	fsys := fs.FS(fstest.MapFS{
		"tmp":            &fstest.MapFile{Mode: fs.ModeDir | 0o777},
		"tmp/foo":        &fstest.MapFile{Data: []byte("foo")},
		"tmp/unreadable": &fstest.MapFile{Data: []byte("foo"), Mode: 0o000},
	})
	fsys = WrapWdFS(fsys)

	t.Setenv("FOO_FILE", "/tmp/foo")
	assert.Equal(t, "foo is foo", ExpandEnvFsys(fsys, "foo is $FOO"))

	t.Setenv("FOO_FILE", "/tmp/missing")
	assert.Equal(t, "empty", ExpandEnvFsys(fsys, "${FOO}empty"))

	fsys = writeOnly(fsys)
	t.Setenv("FOO_FILE", "/tmp/unreadable")
	assert.Equal(t, "", ExpandEnvFsys(fsys, "${FOO}"))
}

// Maybe extract this into a separate package sometime...
// writeOnly - represents a filesystem that's writeable, but read operations fail
func writeOnly(fsys fs.FS) fs.FS {
	return &woFS{fsys}
}

type woFS struct {
	fsys fs.FS
}

func (fsys woFS) Open(name string) (fs.File, error) {
	f, err := fsys.fsys.Open(name)
	return writeOnlyFile(f), err
}

func (fsys woFS) ReadDir(_ string) ([]fs.DirEntry, error) {
	return nil, ErrWriteOnly
}

func (fsys woFS) Stat(_ string) (fs.FileInfo, error) {
	return nil, ErrWriteOnly
}

func writeOnlyFile(f fs.File) fs.File {
	if f == nil {
		return nil
	}

	return &woFile{f}
}

type woFile struct {
	fs.File
}

// Write -
func (f woFile) Write(p []byte) (n int, err error) {
	return hackpadfs.WriteFile(f.File, p)
}

// Read is disabled and returns ErrWriteOnly
func (f woFile) Read([]byte) (n int, err error) {
	return 0, ErrWriteOnly
}

var ErrWriteOnly = errors.New("filesystem is write-only")
