package mountfs

import (
	"errors"
	"github.com/blang/vfs"
	"os"
	filepath "path"
	"strings"
)

// ErrBoundary is returned if an operation
// can not act across filesystem boundaries.
var ErrBoundary = errors.New("Crossing boundary")

// Create a new MountFS based on a root filesystem.
func Create(rootFS vfs.Filesystem) *MountFS {
	return &MountFS{
		rootFS:  rootFS,
		mounts:  make(map[string]vfs.Filesystem),
		parents: make(map[string][]string),
	}
}

// MountFS represents a filesystem build upon a root filesystem
// and multiple filesystems can be mounted inside it.
// In contrast to unix filesystems, the mount path may
// not be a directory or event exist.
//
// Only filesystems with the same path separator are compatible.
// It's not possible to mount a specific source directory, only the
// root of the filesystem can be mounted, use a chroot in this case.
// The resulting filesystem is case-sensitive.
type MountFS struct {
	rootFS  vfs.Filesystem
	mounts  map[string]vfs.Filesystem
	parents map[string][]string
}

// Mount mounts a filesystem on the given path.
// Mounts inside mounts are supported, the longest path match will be taken.
// Mount paths may be overwritten if set on the same path.
// Path `/` can be used to change rootfs.
// Only absolute paths are allowed.
func (fs *MountFS) Mount(mount vfs.Filesystem, path string) error {
	pathSeparator := string(fs.rootFS.PathSeparator())

	// Clean path and make absolute
	path = filepath.Clean(path)
	segm := vfs.SplitPath(path, pathSeparator)
	segm[0] = "" // make absolute
	path = strings.Join(segm, pathSeparator)

	// Change rootfs disabled
	if path == "" {
		fs.rootFS = mount
		return nil
	}

	parent := strings.Join(segm[0:len(segm)-1], pathSeparator)
	if parent == "" {
		parent = "/"
	}
	fs.parents[parent] = append(fs.parents[parent], path)
	fs.mounts[path] = mount
	return nil
}

// PathSeparator returns the path separator
func (fs MountFS) PathSeparator() uint8 {
	return fs.rootFS.PathSeparator()
}

// findMount finds a valid mountpoint for the given path.
// It returns the corresponding filesystem and the path inside of this filesystem.
func findMount(path string, mounts map[string]vfs.Filesystem, fallback vfs.Filesystem, pathSeparator string) (vfs.Filesystem, string) {
	path = filepath.Clean(path)
	segs := vfs.SplitPath(path, pathSeparator)
	l := len(segs)
	for i := l; i > 0; i-- {
		mountPath := strings.Join(segs[0:i], pathSeparator)
		if fs, ok := mounts[mountPath]; ok {
			return fs, "/" + strings.Join(segs[i:l], pathSeparator)
		}
	}
	return fallback, path
}

type innerFile struct {
	vfs.File
	name string
}

// Name returns the full path inside mountfs
func (f innerFile) Name() string {
	return f.name
}

// OpenFile find the mount of the given path and executes OpenFile
// on the corresponding filesystem.
// It wraps the resulting file to return the path inside mountfs on Name()
func (fs MountFS) OpenFile(name string, flag int, perm os.FileMode) (vfs.File, error) {
	mount, innerPath := findMount(name, fs.mounts, fs.rootFS, string(fs.PathSeparator()))
	file, err := mount.OpenFile(innerPath, flag, perm)
	return innerFile{File: file, name: name}, err
}

// Remove removes a file or directory
func (fs MountFS) Remove(name string) error {
	mount, innerPath := findMount(name, fs.mounts, fs.rootFS, string(fs.PathSeparator()))
	return mount.Remove(innerPath)
}

// Rename renames a file.
// Renames across filesystems are not allowed.
func (fs MountFS) Rename(oldpath, newpath string) error {
	oldMount, oldInnerPath := findMount(oldpath, fs.mounts, fs.rootFS, string(fs.PathSeparator()))
	newMount, newInnerPath := findMount(newpath, fs.mounts, fs.rootFS, string(fs.PathSeparator()))
	if oldMount != newMount {
		return ErrBoundary
	}
	return oldMount.Rename(oldInnerPath, newInnerPath)
}

// Mkdir creates a directory
func (fs MountFS) Mkdir(name string, perm os.FileMode) error {
	mount, innerPath := findMount(name, fs.mounts, fs.rootFS, string(fs.PathSeparator()))
	return mount.Mkdir(innerPath, perm)
}

type innerFileInfo struct {
	os.FileInfo
	name string
}

func (fi innerFileInfo) Name() string {
	return fi.name
}

// Stat returns the fileinfo of a file
func (fs MountFS) Stat(name string) (os.FileInfo, error) {
	mount, innerPath := findMount(name, fs.mounts, fs.rootFS, string(fs.PathSeparator()))
	fi, err := mount.Stat(innerPath)
	if innerPath == "/" {
		return innerFileInfo{FileInfo: fi, name: filepath.Base(name)}, err
	}
	return fi, err
}

// Lstat returns the fileinfo of a file or link.
func (fs MountFS) Lstat(name string) (os.FileInfo, error) {
	mount, innerPath := findMount(name, fs.mounts, fs.rootFS, string(fs.PathSeparator()))
	fi, err := mount.Lstat(innerPath)
	if innerPath == "/" {
		return innerFileInfo{FileInfo: fi, name: filepath.Base(name)}, err
	}
	return fi, err
}

// ReadDir reads the directory named by path and returns a list of sorted directory entries.
func (fs MountFS) ReadDir(path string) ([]os.FileInfo, error) {
	path = filepath.Clean(path)
	mount, innerPath := findMount(path, fs.mounts, fs.rootFS, string(fs.PathSeparator()))

	fis, err := mount.ReadDir(innerPath)
	if err != nil {
		return fis, err
	}

	// Add mountpoints
	if childs, ok := fs.parents[path]; ok {
		for _, c := range childs {
			mfi, err := fs.Stat(c)
			if err == nil {
				fis = append(fis, mfi)
			}
		}
	}
	return fis, err
}
