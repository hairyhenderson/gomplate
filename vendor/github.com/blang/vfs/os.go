package vfs

import (
	"io/ioutil"
	"os"
)

// OsFS represents a filesystem backed by the filesystem of the underlying OS.
type OsFS struct{}

// OS returns a filesystem backed by the filesystem of the os. It wraps os.* stdlib operations.
func OS() *OsFS {
	return &OsFS{}
}

// PathSeparator returns the path separator
func (fs OsFS) PathSeparator() uint8 {
	return os.PathSeparator
}

// OpenFile wraps os.OpenFile
func (fs OsFS) OpenFile(name string, flag int, perm os.FileMode) (File, error) {
	return os.OpenFile(name, flag, perm)
}

// Remove wraps os.Remove
func (fs OsFS) Remove(name string) error {
	return os.Remove(name)
}

// Mkdir wraps os.Mkdir
func (fs OsFS) Mkdir(name string, perm os.FileMode) error {
	return os.Mkdir(name, perm)
}

// Rename wraps os.Rename
func (fs OsFS) Rename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}

// Stat wraps os.Stat
func (fs OsFS) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

// Lstat wraps os.Lstat
func (fs OsFS) Lstat(name string) (os.FileInfo, error) {
	return os.Lstat(name)
}

// ReadDir wraps ioutil.ReadDir
func (fs OsFS) ReadDir(path string) ([]os.FileInfo, error) {
	return ioutil.ReadDir(path)
}
