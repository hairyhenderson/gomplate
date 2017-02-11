package vfs

import (
	"errors"
	"io"
	"os"
	"strings"
)

var (
	// ErrIsDirectory is returned if a file is a directory
	ErrIsDirectory = errors.New("Is directory")
	// ErrNotDirectory is returned if a file is not a directory
	ErrNotDirectory = errors.New("Is not a directory")
)

// Filesystem represents an abstract filesystem
type Filesystem interface {
	PathSeparator() uint8
	OpenFile(name string, flag int, perm os.FileMode) (File, error)
	Remove(name string) error
	// RemoveAll(path string) error
	Rename(oldpath, newpath string) error
	Mkdir(name string, perm os.FileMode) error
	// Symlink(oldname, newname string) error
	// TempDir() string
	// Chmod(name string, mode FileMode) error
	// Chown(name string, uid, gid int) error
	Stat(name string) (os.FileInfo, error)
	Lstat(name string) (os.FileInfo, error)
	ReadDir(path string) ([]os.FileInfo, error)
}

// File represents a File with common operations.
// It differs from os.File so e.g. Stat() needs to be called from the Filesystem instead.
//   osfile.Stat() -> filesystem.Stat(file.Name())
type File interface {
	Name() string
	Sync() error
	// Truncate shrinks or extends the size of the File to the specified size.
	Truncate(int64) error
	io.Reader
	io.ReaderAt
	io.Writer
	io.Seeker
	io.Closer
}

// Create creates the named file mode 0666 (before umask) on the given Filesystem,
// truncating it if it already exists.
// The associated file descriptor has mode os.O_RDWR.
// If there is an error, it will be of type *os.PathError.
func Create(fs Filesystem, name string) (File, error) {
	return fs.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
}

// Open opens the named file on the given Filesystem for reading.
// If successful, methods on the returned file can be used for reading.
// The associated file descriptor has mode os.O_RDONLY.
// If there is an error, it will be of type *PathError.
func Open(fs Filesystem, name string) (File, error) {
	return fs.OpenFile(name, os.O_RDONLY, 0)
}

// MkdirAll creates a directory named path on the given Filesystem,
// along with any necessary parents, and returns nil,
// or else returns an error.
// The permission bits perm are used for all
// directories that MkdirAll creates.
// If path is already a directory, MkdirAll does nothing
// and returns nil.
func MkdirAll(fs Filesystem, path string, perm os.FileMode) error {
	if dir, err := fs.Stat(path); err == nil {
		if dir.IsDir() {
			return nil
		}
		return &os.PathError{"mkdir", path, ErrNotDirectory}
	}

	parts := SplitPath(path, string(fs.PathSeparator()))
	if len(parts) > 1 {
		// Create parent
		err := MkdirAll(fs, strings.Join(parts[0:len(parts)-1], string(fs.PathSeparator())), perm)
		if err != nil {
			return err
		}
	}

	// Parent now exists; invoke Mkdir and use its result.
	err := fs.Mkdir(path, perm)
	if err != nil {
		// Handle arguments like "foo/." by
		// double-checking that directory doesn't exist.
		dir, err1 := fs.Lstat(path)
		if err1 == nil && dir.IsDir() {
			return nil
		}
		return err
	}
	return nil
}

// RemoveAll removes path and any children it contains.
// It removes everything it can but returns the first error
// it encounters.  If the path does not exist, RemoveAll
// returns nil.
func RemoveAll(fs Filesystem, path string) error {
	if err := fs.Remove(path); err == nil || os.IsNotExist(err) {
		return nil
	}

	// We could not delete it, so might be a directory
	fis, err := fs.ReadDir(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	// Remove contents & return first error.
	err = nil
	for _, fi := range fis {
		err1 := RemoveAll(fs, path+string(fs.PathSeparator())+fi.Name())
		if err == nil {
			err = err1
		}
	}

	// Remove directory itself.
	err1 := fs.Remove(path)
	if err1 == nil || os.IsNotExist(err1) {
		return nil
	}
	if err == nil {
		err = err1
	}
	return err
}
