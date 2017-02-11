package prefixfs

import (
	"os"
	"reflect"
	"testing"

	"github.com/blang/vfs"
	"github.com/blang/vfs/memfs"
)

const prefixPath = "/prefix"

func prefix(path string) string {
	return prefixPath + "/" + path
}

func rootfs() vfs.Filesystem {
	rfs := memfs.Create()
	rfs.Mkdir(prefixPath, 0777)
	return rfs
}

func TestPathSeparator(t *testing.T) {
	rfs := rootfs()
	fs := Create(rfs, prefixPath)

	if fs.PathSeparator() != rfs.PathSeparator() {
		t.Errorf("fs.PathSeparator() != %v", rfs.PathSeparator())
	}
}

func TestOpenFile(t *testing.T) {
	rfs := rootfs()
	fs := Create(rfs, prefixPath)

	f, err := fs.OpenFile("file", os.O_CREATE, 0666)
	defer f.Close()
	if err != nil {
		t.Errorf("OpenFile: %v", err)
	}

	_, err = rfs.Stat(prefix("file"))
	if os.IsNotExist(err) {
		t.Errorf("root:%v not found (%v)", prefix("file"), err)
	}
}

func TestRemove(t *testing.T) {
	rfs := rootfs()
	fs := Create(rfs, prefixPath)

	f, err := fs.OpenFile("file", os.O_CREATE, 0666)
	defer f.Close()
	if err != nil {
		t.Errorf("OpenFile: %v", err)
	}

	err = fs.Remove("file")
	if err != nil {
		t.Errorf("Remove: %v", err)
	}

	_, err = rfs.Stat(prefix("file"))
	if os.IsExist(err) {
		t.Errorf("root:%v found (%v)", prefix("file"), err)
	}
}

func TestRename(t *testing.T) {
	rfs := rootfs()
	fs := Create(rfs, prefixPath)

	f, err := fs.OpenFile("file", os.O_CREATE, 0666)
	defer f.Close()
	if err != nil {
		t.Errorf("OpenFile: %v", err)
	}

	err = fs.Rename("file", "file2")
	if err != nil {
		t.Errorf("Rename: %v", err)
	}

	_, err = rfs.Stat(prefix("file2"))
	if os.IsNotExist(err) {
		t.Errorf("root:%v not found (%v)", prefix("file2"), err)
	}
}

func TestMkdir(t *testing.T) {
	rfs := rootfs()
	fs := Create(rfs, prefixPath)

	err := fs.Mkdir("dir", 0777)
	if err != nil {
		t.Errorf("Mkdir: %v", err)
	}

	_, err = rfs.Stat(prefix("dir"))
	if os.IsNotExist(err) {
		t.Errorf("root:%v not found (%v)", prefix("dir"), err)
	}
}

func TestStat(t *testing.T) {
	rfs := rootfs()
	fs := Create(rfs, prefixPath)

	f, err := fs.OpenFile("file", os.O_CREATE, 0666)
	defer f.Close()
	if err != nil {
		t.Errorf("OpenFile: %v", err)
	}

	fi, err := fs.Stat("file")
	if os.IsNotExist(err) {
		t.Errorf("fs.Stat: %v", err)
	}

	rfi, err := rfs.Stat(prefix("file"))
	if os.IsNotExist(err) {
		t.Errorf("rfs.Stat: %v", err)
	}

	if fi.Name() != rfi.Name() {
		t.Errorf("FileInfo: Name not equal (fs:%v != root:%v)", fi.Name(), rfi.Name())
	}

	if fi.Size() != rfi.Size() {
		t.Errorf("FileInfo: Size not equal (fs:%v != root:%v)", fi.Size(), rfi.Size())
	}

	if fi.Mode() != rfi.Mode() {
		t.Errorf("FileInfo: Mode not equal (fs:%v != root:%v)", fi.Mode(), rfi.Mode())
	}

	if fi.ModTime() != rfi.ModTime() {
		t.Errorf("FileInfo: ModTime not equal (fs:%v != root:%v)", fi.ModTime(), rfi.ModTime())
	}

	if fi.IsDir() != rfi.IsDir() {
		t.Errorf("FileInfo: Mode not equal (fs:%v != root:%v)", fi.IsDir(), rfi.IsDir())
	}

	if fi.Sys() != rfi.Sys() {
		t.Errorf("FileInfo: Sys not equal (fs:%v != root:%v)", fi.Sys(), rfi.Sys())
	}
}

func TestLstat(t *testing.T) {
	rfs := rootfs()
	fs := Create(rfs, prefixPath)

	f, err := fs.OpenFile("file", os.O_CREATE, 0666)
	defer f.Close()
	if err != nil {
		t.Errorf("OpenFile: %v", err)
	}

	fi, err := fs.Lstat("file")
	if os.IsNotExist(err) {
		t.Errorf("fs.Lstat: %v", err)
	}

	rfi, err := rfs.Lstat(prefix("file"))
	if os.IsNotExist(err) {
		t.Errorf("rfs.Lstat: %v", err)
	}

	if fi.Name() != rfi.Name() {
		t.Errorf("FileInfo: Name not equal (fs:%v != root:%v)", fi.Name(), rfi.Name())
	}

	if fi.Size() != rfi.Size() {
		t.Errorf("FileInfo: Size not equal (fs:%v != root:%v)", fi.Size(), rfi.Size())
	}

	if fi.Mode() != rfi.Mode() {
		t.Errorf("FileInfo: Mode not equal (fs:%v != root:%v)", fi.Mode(), rfi.Mode())
	}

	if fi.ModTime() != rfi.ModTime() {
		t.Errorf("FileInfo: ModTime not equal (fs:%v != root:%v)", fi.ModTime(), rfi.ModTime())
	}

	if fi.IsDir() != rfi.IsDir() {
		t.Errorf("FileInfo: Mode not equal (fs:%v != root:%v)", fi.IsDir(), rfi.IsDir())
	}

	if fi.Sys() != rfi.Sys() {
		t.Errorf("FileInfo: Sys not equal (fs:%v != root:%v)", fi.Sys(), rfi.Sys())
	}
}

func TestReadDir(t *testing.T) {
	rfs := rootfs()
	fs := Create(rfs, prefixPath)

	err := fs.Mkdir("dir", 0777)
	if err != nil {
		t.Errorf("Mkdir: %v", err)
	}

	_, err = rfs.Stat(prefix("dir"))
	if os.IsNotExist(err) {
		t.Errorf("root:%v not found (%v)", prefix("dir"), err)
	}

	f, err := fs.OpenFile("dir/file", os.O_CREATE, 0666)
	defer f.Close()
	if err != nil {
		t.Errorf("OpenFile: %v", err)
	}

	s, err := fs.ReadDir("dir")
	if err != nil {
		t.Errorf("fs.ReadDir: %v", err)
	}

	rs, err := rfs.ReadDir(prefix("dir"))
	if err != nil {
		t.Errorf("rfs.ReadDir: %v", err)
	}

	if !reflect.DeepEqual(s, rs) {
		t.Error("ReadDir: slices not equal")
	}
}
