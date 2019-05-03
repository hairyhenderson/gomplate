package env

import (
	"errors"
	"os"
	"testing"

	"github.com/spf13/afero"

	"github.com/stretchr/testify/assert"
)

func TestGetenv(t *testing.T) {
	assert.Empty(t, Getenv("FOOBARBAZ"))
	assert.Equal(t, os.Getenv("USER"), Getenv("USER"))
	assert.Equal(t, "default value", Getenv("BLAHBLAHBLAH", "default value"))
}

func TestGetenvFile(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = fs.Mkdir("/tmp", 0777)
	f, _ := fs.Create("/tmp/foo")
	_, _ = f.Write([]byte("foo"))

	defer os.Unsetenv("FOO_FILE")
	os.Setenv("FOO_FILE", "/tmp/foo")
	assert.Equal(t, "foo", getenvVFS(fs, "FOO", "bar"))

	os.Setenv("FOO_FILE", "/tmp/missing")
	assert.Equal(t, "bar", getenvVFS(fs, "FOO", "bar"))

	_, _ = fs.Create("/tmp/unreadable")
	fs = writeOnly(fs)
	os.Setenv("FOO_FILE", "/tmp/unreadable")
	assert.Equal(t, "bar", getenvVFS(fs, "FOO", "bar"))
}

func TestExpandEnv(t *testing.T) {
	assert.Empty(t, ExpandEnv("${FOOBARBAZ}"))
	assert.Equal(t, os.Getenv("USER"), ExpandEnv("$USER"))
	assert.Equal(t, "something", ExpandEnv("something$BLAHBLAHBLAH"))
	assert.Equal(t, os.Getenv("USER")+": "+os.Getenv("HOME"),
		ExpandEnv("$USER: ${HOME}"))
}

func TestExpandEnvFile(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = fs.Mkdir("/tmp", 0777)
	f, _ := fs.Create("/tmp/foo")
	_, _ = f.Write([]byte("foo"))

	defer os.Unsetenv("FOO_FILE")
	os.Setenv("FOO_FILE", "/tmp/foo")
	assert.Equal(t, "foo is foo", expandEnvVFS(fs, "foo is $FOO"))

	os.Setenv("FOO_FILE", "/tmp/missing")
	assert.Equal(t, "empty", expandEnvVFS(fs, "${FOO}empty"))

	_, _ = fs.Create("/tmp/unreadable")
	fs = writeOnly(fs)
	os.Setenv("FOO_FILE", "/tmp/unreadable")
	assert.Equal(t, "", expandEnvVFS(fs, "${FOO}"))
}

// Maybe extract this into a separate package sometime...
// writeOnly - represents a filesystem that's writeable, but read operations fail
func writeOnly(fs afero.Fs) afero.Fs {
	return &woFS{fs}
}

type woFS struct {
	afero.Fs
}

func (fs woFS) Remove(name string) error {
	return fs.Fs.Remove(name)
}

func (fs woFS) Rename(oldpath, newpath string) error {
	return fs.Fs.Rename(oldpath, newpath)
}

func (fs woFS) Mkdir(name string, perm os.FileMode) error {
	return fs.Fs.Mkdir(name, perm)
}

func (fs woFS) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	f, err := fs.Fs.OpenFile(name, flag, perm)
	if err != nil {
		return writeOnlyFile(f), err
	}
	return writeOnlyFile(f), nil
}

func (fs woFS) ReadDir(path string) ([]os.FileInfo, error) {
	return nil, ErrWriteOnly
}

func (fs woFS) Stat(name string) (os.FileInfo, error) {
	return nil, ErrWriteOnly
}

func writeOnlyFile(f afero.File) afero.File {
	return &woFile{f}
}

type woFile struct {
	afero.File
}

// Write is disabled and returns ErrWriteOnly
func (f woFile) Write(p []byte) (n int, err error) {
	return f.File.Write(p)
}

// Read is disabled and returns ErrWriteOnly
func (f woFile) Read([]byte) (n int, err error) {
	return 0, ErrWriteOnly
}

var ErrWriteOnly = errors.New("filesystem is write-only")
