package vfs

import (
	"errors"
	"os"
)

// ReadOnly creates a readonly wrapper around the given filesystem.
// It disables the following operations:
//
// 	- Create
// 	- Remove
// 	- Rename
// 	- Mkdir
//
// And disables OpenFile flags: os.O_CREATE, os.O_APPEND, os.O_WRONLY
//
// OpenFile returns a File with disabled Write() method otherwise.
func ReadOnly(fs Filesystem) *RoFS {
	return &RoFS{Filesystem: fs}
}

// RoFS represents a read-only filesystem and
// works as a wrapper around existing filesystems.
type RoFS struct {
	Filesystem
}

// ErrorReadOnly is returned on every disabled operation.
var ErrReadOnly = errors.New("Filesystem is read-only")

// Remove is disabled and returns ErrorReadOnly
func (fs RoFS) Remove(name string) error {
	return ErrReadOnly
}

// Rename is disabled and returns ErrorReadOnly
func (fs RoFS) Rename(oldpath, newpath string) error {
	return ErrReadOnly
}

// Mkdir is disabled and returns ErrorReadOnly
func (fs RoFS) Mkdir(name string, perm os.FileMode) error {
	return ErrReadOnly
}

// OpenFile returns ErrorReadOnly if flag contains os.O_CREATE, os.O_APPEND, os.O_WRONLY.
// Otherwise it returns a read-only File with disabled Write(..) operation.
func (fs RoFS) OpenFile(name string, flag int, perm os.FileMode) (File, error) {
	if flag&os.O_CREATE == os.O_CREATE {
		return nil, ErrReadOnly
	}
	if flag&os.O_APPEND == os.O_APPEND {
		return nil, ErrReadOnly
	}
	if flag&os.O_WRONLY == os.O_WRONLY {
		return nil, ErrReadOnly
	}
	f, err := fs.Filesystem.OpenFile(name, flag, perm)
	if err != nil {
		return ReadOnlyFile(f), err
	}
	return ReadOnlyFile(f), nil
}

// ReadOnlyFile wraps the given file and disables Write(..) operation.
func ReadOnlyFile(f File) File {
	return &roFile{f}
}

type roFile struct {
	File
}

// Write is disabled and returns ErrorReadOnly
func (f roFile) Write(p []byte) (n int, err error) {
	return 0, ErrReadOnly
}
