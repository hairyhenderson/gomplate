package env

import (
	"errors"
	"os"
	"testing"

	"github.com/blang/vfs"
	"github.com/blang/vfs/memfs"
	"github.com/stretchr/testify/assert"
)

func TestGetenv(t *testing.T) {
	assert.Empty(t, Getenv("FOOBARBAZ"))
	assert.Equal(t, os.Getenv("USER"), Getenv("USER"))
	assert.Equal(t, "default value", Getenv("BLAHBLAHBLAH", "default value"))
}

func TestGetenvFile(t *testing.T) {
	var fs vfs.Filesystem
	fs = memfs.Create()
	_ = fs.Mkdir("/tmp", 0777)
	f, _ := vfs.Create(fs, "/tmp/foo")
	_, _ = f.Write([]byte("foo"))

	defer os.Unsetenv("FOO_FILE")
	os.Setenv("FOO_FILE", "/tmp/foo")
	assert.Equal(t, "foo", GetenvVFS(fs, "FOO", "bar"))

	os.Setenv("FOO_FILE", "/tmp/missing")
	assert.Equal(t, "bar", GetenvVFS(fs, "FOO", "bar"))

	f, _ = vfs.Create(fs, "/tmp/unreadable")
	fs = WriteOnly(fs)
	os.Setenv("FOO_FILE", "/tmp/unreadable")
	assert.Equal(t, "bar", GetenvVFS(fs, "FOO", "bar"))
}

func TestExpandEnv(t *testing.T) {
	assert.Empty(t, ExpandEnv("${FOOBARBAZ}"))
	assert.Equal(t, os.Getenv("USER"), ExpandEnv("$USER"))
	assert.Equal(t, "something", ExpandEnv("something$BLAHBLAHBLAH"))
	assert.Equal(t, os.Getenv("USER")+": "+os.Getenv("HOME"),
		ExpandEnv("$USER: ${HOME}"))
}

func TestExpandEnvFile(t *testing.T) {
	var fs vfs.Filesystem
	fs = memfs.Create()
	_ = fs.Mkdir("/tmp", 0777)
	f, _ := vfs.Create(fs, "/tmp/foo")
	_, _ = f.Write([]byte("foo"))

	defer os.Unsetenv("FOO_FILE")
	os.Setenv("FOO_FILE", "/tmp/foo")
	assert.Equal(t, "foo is foo", expandEnvVFS(fs, "foo is $FOO"))

	os.Setenv("FOO_FILE", "/tmp/missing")
	assert.Equal(t, "empty", expandEnvVFS(fs, "${FOO}empty"))

	f, _ = vfs.Create(fs, "/tmp/unreadable")
	fs = WriteOnly(fs)
	os.Setenv("FOO_FILE", "/tmp/unreadable")
	assert.Equal(t, "", expandEnvVFS(fs, "${FOO}"))
}

// Maybe extract this into a separate package sometime...
// WriteOnly - represents a filesystem that's writeable, but read operations fail
func WriteOnly(fs vfs.Filesystem) vfs.Filesystem {
	return &WoFS{fs}
}

type WoFS struct {
	vfs.Filesystem
}

func (fs WoFS) Remove(name string) error {
	return fs.Filesystem.Remove(name)
}

func (fs WoFS) Rename(oldpath, newpath string) error {
	return fs.Filesystem.Rename(oldpath, newpath)
}

func (fs WoFS) Mkdir(name string, perm os.FileMode) error {
	return fs.Filesystem.Mkdir(name, perm)
}

func (fs WoFS) OpenFile(name string, flag int, perm os.FileMode) (vfs.File, error) {
	f, err := fs.Filesystem.OpenFile(name, flag, perm)
	if err != nil {
		return WriteOnlyFile(f), err
	}
	return WriteOnlyFile(f), nil
}

func (fs WoFS) Lstat(name string) (os.FileInfo, error) {
	return fs.Filesystem.Lstat(name)
}

func (fs WoFS) PathSeparator() uint8 {
	return fs.Filesystem.PathSeparator()
}

func (fs WoFS) ReadDir(path string) ([]os.FileInfo, error) {
	return nil, ErrWriteOnly
}

func (fs WoFS) Stat(name string) (os.FileInfo, error) {
	return nil, ErrWriteOnly
}

func WriteOnlyFile(f vfs.File) vfs.File {
	return &woFile{f}
}

type woFile struct {
	vfs.File
}

// Write is disabled and returns ErrWriteOnly
func (f woFile) Write(p []byte) (n int, err error) {
	return f.File.Write(p)
}

// Read is disabled and returns ErrWriteOnly
func (f woFile) Read([]byte) (n int, err error) {
	return 0, ErrWriteOnly
}

var ErrWriteOnly = errors.New("Filesystem is write-only")
