package vfs

import (
	"errors"
	"os"
	"testing"
)

type openFS struct {
	Filesystem
	fn func(name string, flag int, perm os.FileMode) (File, error)
}

func (fs openFS) OpenFile(name string, flag int, perm os.FileMode) (File, error) {
	return fs.fn(name, flag, perm)
}

func TestCreate(t *testing.T) {
	fs := openFS{
		Filesystem: Dummy(errors.New("Not implemented")),
		fn: func(name string, flag int, perm os.FileMode) (File, error) {
			if name != "name" {
				t.Errorf("Invalid name: %s", name)
			}
			if flag != os.O_RDWR|os.O_CREATE|os.O_TRUNC {
				t.Errorf("Invalid flag: %d", flag)
			}
			if perm != 0666 {
				t.Errorf("Invalid perm: %d", perm)
			}
			return nil, nil
		},
	}
	_, err := Create(fs, "name")
	if err != nil {
		t.Fatalf("OpenFile not called")
	}
}

func TestOpen(t *testing.T) {
	fs := openFS{
		Filesystem: Dummy(errors.New("Not implemented")),
		fn: func(name string, flag int, perm os.FileMode) (File, error) {
			if name != "name" {
				t.Errorf("Invalid name: %s", name)
			}
			if flag != os.O_RDONLY {
				t.Errorf("Invalid flag: %d", flag)
			}
			if perm != 0 {
				t.Errorf("Invalid perm: %d", perm)
			}
			return nil, nil
		},
	}
	_, err := Open(fs, "name")
	if err != nil {
		t.Fatalf("OpenFile not called")
	}
}

type mkdirFS struct {
	Filesystem
	dirs map[string]os.FileInfo
	perm os.FileMode
}

func (fs *mkdirFS) Mkdir(name string, perm os.FileMode) error {
	if _, ok := fs.dirs[name]; ok {
		return os.ErrExist
	}
	fs.perm = perm
	fs.dirs[name] = DumFileInfo{
		IName: name,
		IDir:  true,
		IMode: perm,
	}
	return nil
}

func (fs mkdirFS) Stat(name string) (os.FileInfo, error) {
	return fs.Lstat(name)
}

func (fs mkdirFS) Lstat(name string) (os.FileInfo, error) {
	if fi, ok := fs.dirs[name]; ok {
		return fi, nil
	}
	return nil, os.ErrNotExist
}

func TestMkdirAll(t *testing.T) {
	fs := &mkdirFS{
		Filesystem: Dummy(errors.New("Not implemented")),
		dirs:       make(map[string]os.FileInfo),
	}
	err := MkdirAll(fs, "/usr/src/linux", 0777)
	if err != nil {
		t.Fatalf("Mkdir failed")
	}
	if fs.perm != 0777 {
		t.Errorf("Wrong perm: %d", fs.perm)
	}
	if _, ok := fs.dirs["/usr"]; !ok {
		t.Errorf("Dir not created: /usr")
	}
	if _, ok := fs.dirs["/usr/src"]; !ok {
		t.Errorf("Dir not created: /usr/src")
	}
	if _, ok := fs.dirs["/usr/src/linux"]; !ok {
		t.Errorf("Dir not created: /usr/src/linux")
	}
}

func TestMkdirAllExists(t *testing.T) {
	fs := &mkdirFS{
		Filesystem: Dummy(errors.New("Not implemented")),
		dirs:       make(map[string]os.FileInfo),
	}
	// Make dir
	fs.dirs["/usr/src/linux"] = DumFileInfo{IName: "linux", IDir: true}

	err := MkdirAll(fs, "/usr/src/linux", 0777)
	if err != nil {
		t.Fatalf("Mkdir failed")
	}
}

func TestMkdirAllFirstExists(t *testing.T) {
	fs := &mkdirFS{
		Filesystem: Dummy(errors.New("Not implemented")),
		dirs:       make(map[string]os.FileInfo),
	}
	// Make dir
	fs.dirs["/usr"] = DumFileInfo{IName: "usr", IDir: true}

	err := MkdirAll(fs, "/usr/src/linux/", 0777)
	if err != nil {
		t.Fatalf("Mkdir failed")
	}

	if _, ok := fs.dirs["/usr/src"]; !ok {
		t.Errorf("Dir not created: /usr/src")
	}
	if _, ok := fs.dirs["/usr/src/linux/"]; !ok {
		t.Errorf("Dir not created: /usr/src/linux")
	}
}

func TestMkdirAllFirstExistsNoFile(t *testing.T) {
	fs := &mkdirFS{
		Filesystem: Dummy(errors.New("Not implemented")),
		dirs:       make(map[string]os.FileInfo),
	}
	// Make dir
	fs.dirs["/usr"] = DumFileInfo{IName: "usr", IDir: true}
	fs.dirs["/usr/src"] = DumFileInfo{IName: "src", IDir: false}

	err := MkdirAll(fs, "/usr/src/linux/linux-4.10", 0777)
	if err == nil {
		t.Fatalf("Mkdir failed")
	}
}

type rmFileinfo struct {
	DumFileInfo
	subfiles []*rmFileinfo
	parent   *rmFileinfo
}

type rmFS struct {
	Filesystem
	files map[string]*rmFileinfo
}

func (fs *rmFS) ReadDir(path string) ([]os.FileInfo, error) {
	if fi, ok := fs.files[path]; ok {
		if fi.IsDir() {
			s := make([]os.FileInfo, len(fi.subfiles))
			for i, sf := range fi.subfiles {

				s[i] = sf
			}
			for _, sf := range s {
				if sf == nil {
					panic("sf in readdir nil")
				}
			}
			return s, nil
		}
		return nil, ErrNotDirectory
	}

	return nil, os.ErrNotExist
}

func findRmFileInfoIndex(s []*rmFileinfo, needle *rmFileinfo) int {
	for i, fi := range s {
		if fi == needle {
			return i
		}
	}
	return -1
}

func (fs *rmFS) Remove(name string) error {
	if fi, ok := fs.files[name]; ok {
		if fi.IsDir() && len(fi.subfiles) > 0 {
			return ErrIsDirectory // Not empty
		}

		// remove references
		delete(fs.files, name)
		if fi.parent != nil {
			if i := findRmFileInfoIndex(fi.parent.subfiles, fi); i >= 0 {
				fi.parent.subfiles = append(fi.parent.subfiles[:i], fi.parent.subfiles[i+1:]...)
			}
		}
		return nil
	}

	return &os.PathError{"remove", name, os.ErrNotExist}
}

func TestRemoveAll(t *testing.T) {
	fs := &rmFS{
		Filesystem: Dummy(errors.New("Not implemented")),
		files:      make(map[string]*rmFileinfo),
	}

	fiTmpFile := &rmFileinfo{
		DumFileInfo: DumFileInfo{
			IName: "file",
			IDir:  false,
		},
	}

	fiTmp := &rmFileinfo{
		DumFileInfo: DumFileInfo{
			IName: "tmp",
			IDir:  true,
		},
		subfiles: []*rmFileinfo{
			fiTmpFile,
		},
	}

	fiRoot := &rmFileinfo{
		DumFileInfo: DumFileInfo{
			IName: "/",
			IDir:  true,
		},
		subfiles: []*rmFileinfo{
			fiTmp,
		},
	}
	fs.files["/tmp/file"] = fiTmpFile
	fiTmpFile.parent = fiTmp
	fs.files["/tmp"] = fiTmp
	fiTmp.parent = fiRoot
	fs.files["/"] = fiRoot

	fiTmpFile.Name()
	err := RemoveAll(fs, "/tmp")
	if err != nil {
		t.Errorf("Unexpected error remove all: %s", err)
	}

	if _, ok := fs.files["/tmp/file"]; ok {
		t.Errorf("/tmp/file was not removed")
	}

	if _, ok := fs.files["/tmp"]; ok {
		t.Errorf("/tmp was not removed")
	}

	if _, ok := fs.files["/"]; !ok {
		t.Errorf("/ was removed")
	}

}
