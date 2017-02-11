package prefixfs

import (
	"os"

	"github.com/blang/vfs"
)

// A FS that prefixes the path in each vfs.Filesystem operation.
type FS struct {
	vfs.Filesystem

	// Prefix is used to prefix the path in each vfs.Filesystem operation.
	Prefix string
}

// Create returns a file system that prefixes all paths and forwards to root.
func Create(root vfs.Filesystem, prefix string) *FS {
	return &FS{root, prefix}
}

// PrefixPath returns path with the prefix prefixed.
func (fs *FS) PrefixPath(path string) string {
	return fs.Prefix + string(fs.PathSeparator()) + path
}

// PathSeparator implements vfs.Filesystem.
func (fs *FS) PathSeparator() uint8 { return fs.Filesystem.PathSeparator() }

// OpenFile implements vfs.Filesystem.
func (fs *FS) OpenFile(name string, flag int, perm os.FileMode) (vfs.File, error) {
	return fs.Filesystem.OpenFile(fs.PrefixPath(name), flag, perm)
}

// Remove implements vfs.Filesystem.
func (fs *FS) Remove(name string) error {
	return fs.Filesystem.Remove(fs.PrefixPath(name))
}

// Rename implements vfs.Filesystem.
func (fs *FS) Rename(oldpath, newpath string) error {
	return fs.Filesystem.Rename(fs.PrefixPath(oldpath), fs.PrefixPath(newpath))
}

// Mkdir implements vfs.Filesystem.
func (fs *FS) Mkdir(name string, perm os.FileMode) error {
	return fs.Filesystem.Mkdir(fs.PrefixPath(name), perm)
}

// Stat implements vfs.Filesystem.
func (fs *FS) Stat(name string) (os.FileInfo, error) {
	return fs.Filesystem.Stat(fs.PrefixPath(name))
}

// Lstat implements vfs.Filesystem.
func (fs *FS) Lstat(name string) (os.FileInfo, error) {
	return fs.Filesystem.Lstat(fs.PrefixPath(name))
}

// ReadDir implements vfs.Filesystem.
func (fs *FS) ReadDir(path string) ([]os.FileInfo, error) {
	return fs.Filesystem.ReadDir(fs.PrefixPath(path))
}
