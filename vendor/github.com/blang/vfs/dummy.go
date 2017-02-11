package vfs

import (
	"os"
	"time"
)

// Dummy creates a new dummy filesystem which returns the given error on every operation.
func Dummy(err error) *DummyFS {
	return &DummyFS{err}
}

// DummyFS is dummy filesystem which returns an error on every operation.
// It can be used to mock a full filesystem for testing or fs creation.
type DummyFS struct {
	err error
}

// PathSeparator returns the path separator
func (fs DummyFS) PathSeparator() uint8 {
	return '/'
}

// OpenFile returns dummy error
func (fs DummyFS) OpenFile(name string, flag int, perm os.FileMode) (File, error) {
	return nil, fs.err
}

// Remove returns dummy error
func (fs DummyFS) Remove(name string) error {
	return fs.err
}

// Rename returns dummy error
func (fs DummyFS) Rename(oldpath, newpath string) error {
	return fs.err
}

// Mkdir returns dummy error
func (fs DummyFS) Mkdir(name string, perm os.FileMode) error {
	return fs.err
}

// Stat returns dummy error
func (fs DummyFS) Stat(name string) (os.FileInfo, error) {
	return nil, fs.err
}

// Lstat returns dummy error
func (fs DummyFS) Lstat(name string) (os.FileInfo, error) {
	return nil, fs.err
}

// ReadDir returns dummy error
func (fs DummyFS) ReadDir(path string) ([]os.FileInfo, error) {
	return nil, fs.err
}

// DummyFile mocks a File returning an error on every operation
// To create a DummyFS returning a dummyFile instead of an error
// you can your own DummyFS:
//
// 	type writeDummyFS struct {
// 		Filesystem
// 	}
//
// 	func (fs writeDummyFS) OpenFile(name string, flag int, perm os.FileMode) (File, error) {
// 		return DummyFile(dummyError), nil
// 	}
func DummyFile(err error) *DumFile {
	return &DumFile{err}
}

// DumFile represents a dummy File
type DumFile struct {
	err error
}

// Name returns "dummy"
func (f DumFile) Name() string {
	return "dummy"
}

// Sync returns dummy error
func (f DumFile) Sync() error {
	return f.err
}

// Truncate returns dummy error
func (f DumFile) Truncate(size int64) error {
	return f.err
}

// Close returns dummy error
func (f DumFile) Close() error {
	return f.err
}

// Write returns dummy error
func (f DumFile) Write(p []byte) (n int, err error) {
	return 0, f.err
}

// Read returns dummy error
func (f DumFile) Read(p []byte) (n int, err error) {
	return 0, f.err
}

// ReadAt returns dummy error
func (f DumFile) ReadAt(p []byte, off int64) (n int, err error) {
	return 0, f.err
}

// Seek returns dummy error
func (f DumFile) Seek(offset int64, whence int) (int64, error) {
	return 0, f.err
}

// DumFileInfo mocks a os.FileInfo returning default values on every operation
// Struct fields can be set.
type DumFileInfo struct {
	IName    string
	ISize    int64
	IMode    os.FileMode
	IModTime time.Time
	IDir     bool
	ISys     interface{}
}

// Name returns the field IName
func (fi DumFileInfo) Name() string {
	return fi.IName
}

// Size returns the field ISize
func (fi DumFileInfo) Size() int64 {
	return fi.ISize
}

// Mode returns the field IMode
func (fi DumFileInfo) Mode() os.FileMode {
	return fi.IMode
}

// ModTime returns the field IModTime
func (fi DumFileInfo) ModTime() time.Time {
	return fi.IModTime
}

// IsDir returns the field IDir
func (fi DumFileInfo) IsDir() bool {
	return fi.IDir
}

// Sys returns the field ISys
func (fi DumFileInfo) Sys() interface{} {
	return fi.ISys
}
