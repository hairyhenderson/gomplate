package memfs

import (
	"errors"
	"fmt"
	"os"
	filepath "path"
	"sort"
	"sync"
	"time"

	"github.com/blang/vfs"
)

var (
	// ErrReadOnly is returned if the file is read-only and write operations are disabled.
	ErrReadOnly = errors.New("File is read-only")
	// ErrWriteOnly is returned if the file is write-only and read operations are disabled.
	ErrWriteOnly = errors.New("File is write-only")
	// ErrIsDirectory is returned if the file under operation is not a regular file but a directory.
	ErrIsDirectory = errors.New("Is directory")
)

// PathSeparator used to separate path segments
const PathSeparator = "/"

// MemFS is a in-memory filesystem
type MemFS struct {
	root *fileInfo
	wd   *fileInfo
	lock *sync.RWMutex
}

// Create a new MemFS filesystem which entirely resides in memory
func Create() *MemFS {
	root := &fileInfo{
		name: "/",
		dir:  true,
	}
	return &MemFS{
		root: root,
		wd:   root,
		lock: &sync.RWMutex{},
	}
}

type fileInfo struct {
	name    string
	dir     bool
	mode    os.FileMode
	parent  *fileInfo
	size    int64
	modTime time.Time
	fs      vfs.Filesystem
	childs  map[string]*fileInfo
	buf     *[]byte
	mutex   *sync.RWMutex
}

func (fi fileInfo) Sys() interface{} {
	return fi.fs
}

func (fi fileInfo) Size() int64 {
	if fi.dir {
		return 0
	}
	fi.mutex.RLock()
	l := len(*(fi.buf))
	fi.mutex.RUnlock()
	return int64(l)
}

func (fi fileInfo) IsDir() bool {
	return fi.dir
}

// ModTime returns the modification time.
// Modification time is updated on:
// 	- Creation
// 	- Rename
// 	- Open (except with O_RDONLY)
func (fi fileInfo) ModTime() time.Time {
	return fi.modTime
}

func (fi fileInfo) Mode() os.FileMode {
	return fi.mode
}

func (fi fileInfo) Name() string {
	return fi.name
}

func (fi fileInfo) AbsPath() string {
	if fi.parent != nil {
		return filepath.Join(fi.parent.AbsPath(), fi.name)
	}
	return "/"
}

// PathSeparator returns the path separator
func (fs *MemFS) PathSeparator() uint8 {
	return '/'
}

// Mkdir creates a new directory with given permissions
func (fs *MemFS) Mkdir(name string, perm os.FileMode) error {
	fs.lock.Lock()
	defer fs.lock.Unlock()
	name = filepath.Clean(name)
	base := filepath.Base(name)
	parent, fi, err := fs.fileInfo(name)
	if err != nil {
		return &os.PathError{"mkdir", name, err}
	}
	if fi != nil {
		return &os.PathError{"mkdir", name, fmt.Errorf("Directory %q already exists", name)}
	}

	fi = &fileInfo{
		name:    base,
		dir:     true,
		mode:    perm,
		parent:  parent,
		modTime: time.Now(),
		fs:      fs,
	}
	parent.childs[base] = fi
	return nil
}

// byName implements sort.Interface
type byName []os.FileInfo

// Len returns the length of the slice
func (f byName) Len() int { return len(f) }

// Less sorts slice by Name
func (f byName) Less(i, j int) bool { return f[i].Name() < f[j].Name() }

// Swap two elements by index
func (f byName) Swap(i, j int) { f[i], f[j] = f[j], f[i] }

// ReadDir reads the directory named by path and returns a list of sorted directory entries.
func (fs *MemFS) ReadDir(path string) ([]os.FileInfo, error) {
	fs.lock.RLock()
	defer fs.lock.RUnlock()

	path = filepath.Clean(path)
	_, fi, err := fs.fileInfo(path)
	if err != nil {
		return nil, &os.PathError{"readdir", path, err}
	}
	if fi == nil || !fi.dir {
		return nil, &os.PathError{"readdir", path, vfs.ErrNotDirectory}
	}

	fis := make([]os.FileInfo, 0, len(fi.childs))
	for _, e := range fi.childs {
		fis = append(fis, e)
	}
	sort.Sort(byName(fis))
	return fis, nil
}

func (fs *MemFS) fileInfo(path string) (parent *fileInfo, node *fileInfo, err error) {
	path = filepath.Clean(path)
	segments := vfs.SplitPath(path, PathSeparator)

	// Shortcut for working directory and root
	if len(segments) == 1 {
		if segments[0] == "" {
			return nil, fs.root, nil
		} else if segments[0] == "." {
			return fs.wd.parent, fs.wd, nil
		}
	}

	// Determine root to traverse
	parent = fs.root
	if segments[0] == "." {
		parent = fs.wd
	}
	segments = segments[1:]

	// Further directories
	if len(segments) > 1 {
		for _, seg := range segments[:len(segments)-1] {

			if parent.childs == nil {
				return nil, nil, os.ErrNotExist
			}
			if entry, ok := parent.childs[seg]; ok && entry.dir {
				parent = entry
			} else {
				return nil, nil, os.ErrNotExist
			}
		}
	}

	lastSeg := segments[len(segments)-1]
	if parent.childs != nil {
		if node, ok := parent.childs[lastSeg]; ok {
			return parent, node, nil
		}
	} else {
		parent.childs = make(map[string]*fileInfo)
	}

	return parent, nil, nil
}

func hasFlag(flag int, flags int) bool {
	return flags&flag == flag
}

// OpenFile opens a file handle with a specified flag (os.O_RDONLY etc.) and perm (e.g. 0666).
// If success the returned File can be used for I/O. Otherwise an error is returned, which
// is a *os.PathError and can be extracted for further information.
func (fs *MemFS) OpenFile(name string, flag int, perm os.FileMode) (vfs.File, error) {
	fs.lock.Lock()
	defer fs.lock.Unlock()

	name = filepath.Clean(name)
	base := filepath.Base(name)
	fiParent, fiNode, err := fs.fileInfo(name)
	if err != nil {
		return nil, &os.PathError{"open", name, err}
	}

	if fiNode == nil {
		if !hasFlag(os.O_CREATE, flag) {
			return nil, &os.PathError{"open", name, os.ErrNotExist}
		}
		fiNode = &fileInfo{
			name:    base,
			dir:     false,
			mode:    perm,
			parent:  fiParent,
			modTime: time.Now(),
			fs:      fs,
		}
		fiParent.childs[base] = fiNode
	} else { // file exists
		if hasFlag(os.O_CREATE|os.O_EXCL, flag) {
			return nil, &os.PathError{"open", name, os.ErrExist}
		}
		if fiNode.dir {
			return nil, &os.PathError{"open", name, ErrIsDirectory}
		}
	}

	if !hasFlag(os.O_RDONLY, flag) {
		fiNode.modTime = time.Now()
	}
	return fiNode.file(flag)
}

func (fi *fileInfo) file(flag int) (vfs.File, error) {
	if fi.buf == nil || hasFlag(os.O_TRUNC, flag) {
		buf := make([]byte, 0, MinBufferSize)
		fi.buf = &buf
		fi.mutex = &sync.RWMutex{}
	}
	var f vfs.File = NewMemFile(fi.AbsPath(), fi.mutex, fi.buf)
	if hasFlag(os.O_APPEND, flag) {
		f.Seek(0, os.SEEK_END)
	}
	if hasFlag(os.O_RDWR, flag) {
		return f, nil
	} else if hasFlag(os.O_WRONLY, flag) {
		f = &woFile{f}
	} else {
		f = &roFile{f}
	}

	return f, nil
}

// roFile wraps the given file and disables Write(..) operation.
type roFile struct {
	vfs.File
}

// Write is disabled and returns ErrorReadOnly
func (f *roFile) Write(p []byte) (n int, err error) {
	return 0, ErrReadOnly
}

// woFile wraps the given file and disables Read(..) operation.
type woFile struct {
	vfs.File
}

// Read is disabled and returns ErrorWroteOnly
func (f *woFile) Read(p []byte) (n int, err error) {
	return 0, ErrWriteOnly
}

// Remove removes the named file or directory.
// If there is an error, it will be of type *PathError.
func (fs *MemFS) Remove(name string) error {
	fs.lock.Lock()
	defer fs.lock.Unlock()

	name = filepath.Clean(name)
	fiParent, fiNode, err := fs.fileInfo(name)
	if err != nil {
		return &os.PathError{"remove", name, err}
	}
	if fiNode == nil {
		return &os.PathError{"remove", name, os.ErrNotExist}
	}

	delete(fiParent.childs, fiNode.name)
	return nil
}

// Rename renames (moves) a file.
// Handles to the oldpath persist but might return oldpath if Name() is called.
func (fs *MemFS) Rename(oldpath, newpath string) error {
	fs.lock.Lock()
	defer fs.lock.Unlock()

	// OldPath
	oldpath = filepath.Clean(oldpath)
	fiOldParent, fiOld, err := fs.fileInfo(oldpath)
	if err != nil {
		return &os.PathError{"rename", oldpath, err}
	}
	if fiOld == nil {
		return &os.PathError{"rename", oldpath, os.ErrNotExist}
	}

	newpath = filepath.Clean(newpath)
	fiNewParent, fiNew, err := fs.fileInfo(newpath)
	if err != nil {
		return &os.PathError{"rename", newpath, err}
	}

	if fiNew != nil {
		return &os.PathError{"rename", newpath, os.ErrExist}
	}

	newBase := filepath.Base(newpath)

	// Relink
	delete(fiOldParent.childs, fiOld.name)
	fiOld.parent = fiNewParent
	fiOld.name = newBase
	fiOld.modTime = time.Now()
	fiNewParent.childs[fiOld.name] = fiOld
	return nil
}

// Stat returns the FileInfo structure describing the named file.
// If there is an error, it will be of type *PathError.
func (fs *MemFS) Stat(name string) (os.FileInfo, error) {
	fs.lock.RLock()
	defer fs.lock.RUnlock()

	name = filepath.Clean(name)
	// dir, base := filepath.Split(name)
	_, fi, err := fs.fileInfo(name)
	if err != nil {
		return nil, &os.PathError{"stat", name, err}
	}
	if fi == nil {
		return nil, &os.PathError{"stat", name, os.ErrNotExist}
	}
	return fi, nil
}

// Lstat returns a FileInfo describing the named file.
// MemFS does not support symbolic links.
// Alias for fs.Stat(name)
func (fs *MemFS) Lstat(name string) (os.FileInfo, error) {
	return fs.Stat(name)
}
