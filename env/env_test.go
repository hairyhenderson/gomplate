package env

import (
	"errors"
	"io/fs"
	"os"
	"testing"
	"testing/fstest"

	"github.com/hack-pad/hackpadfs"
	"github.com/hairyhenderson/gomplate/v4/internal/datafs"

	"github.com/stretchr/testify/assert"
)

func TestGetenv(t *testing.T) {
	assert.Empty(t, Getenv("FOOBARBAZ"))
	assert.Equal(t, os.Getenv("USER"), Getenv("USER"))
	assert.Equal(t, "default value", Getenv("BLAHBLAHBLAH", "default value"))
}

func TestGetenvFile(t *testing.T) {
	fsys := fs.FS(fstest.MapFS{
		"tmp":            &fstest.MapFile{Mode: fs.ModeDir | 0o777},
		"tmp/foo":        &fstest.MapFile{Data: []byte("foo")},
		"tmp/unreadable": &fstest.MapFile{Data: []byte("foo"), Mode: 0o000},
	})
	fsys = datafs.WrapWdFS(fsys)

	t.Setenv("FOO_FILE", "/tmp/foo")
	assert.Equal(t, "foo", getenvVFS(fsys, "FOO", "bar"))

	t.Setenv("FOO_FILE", "/tmp/missing")
	assert.Equal(t, "bar", getenvVFS(fsys, "FOO", "bar"))

	fsys = writeOnly(fsys)
	t.Setenv("FOO_FILE", "/tmp/unreadable")
	assert.Equal(t, "bar", getenvVFS(fsys, "FOO", "bar"))
}

func TestExpandEnv(t *testing.T) {
	assert.Empty(t, ExpandEnv("${FOOBARBAZ}"))
	assert.Equal(t, os.Getenv("USER"), ExpandEnv("$USER"))
	assert.Equal(t, "something", ExpandEnv("something$BLAHBLAHBLAH"))
	assert.Equal(t, os.Getenv("USER")+": "+os.Getenv("HOME"),
		ExpandEnv("$USER: ${HOME}"))
}

func TestExpandEnvFile(t *testing.T) {
	fsys := fs.FS(fstest.MapFS{
		"tmp":            &fstest.MapFile{Mode: fs.ModeDir | 0o777},
		"tmp/foo":        &fstest.MapFile{Data: []byte("foo")},
		"tmp/unreadable": &fstest.MapFile{Data: []byte("foo"), Mode: 0o000},
	})
	fsys = datafs.WrapWdFS(fsys)

	t.Setenv("FOO_FILE", "/tmp/foo")
	assert.Equal(t, "foo is foo", expandEnvVFS(fsys, "foo is $FOO"))

	t.Setenv("FOO_FILE", "/tmp/missing")
	assert.Equal(t, "empty", expandEnvVFS(fsys, "${FOO}empty"))

	fsys = writeOnly(fsys)
	t.Setenv("FOO_FILE", "/tmp/unreadable")
	assert.Equal(t, "", expandEnvVFS(fsys, "${FOO}"))
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
