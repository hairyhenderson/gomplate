package memfs

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/blang/vfs"
)

func TestInterface(t *testing.T) {
	_ = vfs.Filesystem(Create())
}

func TestCreate(t *testing.T) {
	fs := Create()
	// Create file with absolute path
	{
		f, err := fs.OpenFile("/testfile", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			t.Fatalf("Unexpected error creating file: %s", err)
		}
		if name := f.Name(); name != "/testfile" {
			t.Errorf("Wrong name: %s", name)
		}
	}

	// Create same file twice, no error because os.O_TRUNC is used
	{
		_, err := fs.OpenFile("/testfile", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			t.Fatalf("Expected error creating file: %s", err)
		}

	}

	// Create file with unkown parent
	{
		_, err := fs.OpenFile("/testfile/testfile", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
		if err == nil {
			t.Errorf("Expected error creating file")
		}
	}

	// Create file with relative path (workingDir == root)
	{
		f, err := fs.OpenFile("relFile", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			t.Fatalf("Unexpected error creating file: %s", err)
		}
		if name := f.Name(); name != "/relFile" {
			t.Errorf("Wrong name: %s", name)
		}
	}

}

func TestMkdirAbsRel(t *testing.T) {
	fs := Create()

	// Create dir with absolute path
	{
		err := fs.Mkdir("/usr", 0)
		if err != nil {
			t.Fatalf("Unexpected error creating directory: %s", err)
		}
	}

	// Create dir with relative path
	{
		err := fs.Mkdir("home", 0)
		if err != nil {
			t.Fatalf("Unexpected error creating directory: %s", err)
		}
	}

	// Create dir twice
	{
		err := fs.Mkdir("/home", 0)
		if err == nil {
			t.Fatalf("Expecting error creating directory: %s", "/home")
		}
	}
}

func TestMkdirTree(t *testing.T) {
	fs := Create()

	err := fs.Mkdir("/home", 0)
	if err != nil {
		t.Fatalf("Unexpected error creating directory /home: %s", err)
	}

	err = fs.Mkdir("/home/blang", 0)
	if err != nil {
		t.Fatalf("Unexpected error creating directory /home/blang: %s", err)
	}

	err = fs.Mkdir("/home/blang/goprojects", 0)
	if err != nil {
		t.Fatalf("Unexpected error creating directory /home/blang/goprojects: %s", err)
	}

	err = fs.Mkdir("/home/johndoe/goprojects", 0)
	if err == nil {
		t.Errorf("Expected error creating directory with non-existing parent")
	}

	//TODO: Subdir of file
}

func TestReadDir(t *testing.T) {
	fs := Create()
	dirs := []string{"/home", "/home/linus", "/home/rob", "/home/pike", "/home/blang"}
	expectNames := []string{"README.txt", "blang", "linus", "pike", "rob"}
	for _, dir := range dirs {
		err := fs.Mkdir(dir, 0777)
		if err != nil {
			t.Fatalf("Unexpected error creating directory %q: %s", dir, err)
		}
	}
	f, err := fs.OpenFile("/home/README.txt", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		t.Fatalf("Unexpected error creating file: %s", err)
	}
	f.Close()

	fis, err := fs.ReadDir("/home")
	if err != nil {
		t.Fatalf("Unexpected error readdir: %s", err)
	}
	if l := len(fis); l != len(expectNames) {
		t.Errorf("Wrong size: %q (%d)", fis, l)
	}
	for i, n := range expectNames {
		if fn := fis[i].Name(); fn != n {
			t.Errorf("Expected name %q, got %q", n, fn)
		}
	}

	// Readdir empty directory
	if fis, err := fs.ReadDir("/home/blang"); err != nil {
		t.Errorf("Error readdir(empty directory): %s", err)
	} else if l := len(fis); l > 0 {
		t.Errorf("Found entries in non-existing directory: %d", l)
	}

	// Readdir file
	if _, err := fs.ReadDir("/home/README.txt"); err == nil {
		t.Errorf("Expected error readdir(file)")
	}

	// Readdir subdir of file
	if _, err := fs.ReadDir("/home/README.txt/info"); err == nil {
		t.Errorf("Expected error readdir(subdir of file)")
	}

	// Readdir non existing directory
	if _, err := fs.ReadDir("/usr"); err == nil {
		t.Errorf("Expected error readdir(nofound)")
	}

}

func TestRemove(t *testing.T) {
	fs := Create()
	err := fs.Mkdir("/tmp", 0777)
	if err != nil {
		t.Fatalf("Mkdir error: %s", err)
	}
	f, err := fs.OpenFile("/tmp/README.txt", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		t.Fatalf("Create error: %s", err)
	}
	if _, err := f.Write([]byte("test")); err != nil {
		t.Fatalf("Write error: %s", err)
	}
	f.Close()

	// remove non existing file
	if err := fs.Remove("/nonexisting.txt"); err == nil {
		t.Errorf("Expected remove to fail")
	}

	// remove non existing file from an non existing directory
	if err := fs.Remove("/nonexisting/nonexisting.txt"); err == nil {
		t.Errorf("Expected remove to fail")
	}

	// remove created file
	err = fs.Remove(f.Name())
	if err != nil {
		t.Errorf("Remove failed: %s", err)
	}

	if _, err = fs.OpenFile("/tmp/README.txt", os.O_RDWR, 0666); err == nil {
		t.Errorf("Could open removed file!")
	}

	err = fs.Remove("/tmp")
	if err != nil {
		t.Errorf("Remove failed: %s", err)
	}
	if fis, err := fs.ReadDir("/"); err != nil {
		t.Errorf("Readdir error: %s", err)
	} else if len(fis) != 0 {
		t.Errorf("Found files: %s", fis)
	}
}

func TestReadWrite(t *testing.T) {
	fs := Create()
	f, err := fs.OpenFile("/readme.txt", os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		t.Fatalf("Could not open file: %s", err)
	}

	// Write first dots
	if n, err := f.Write([]byte(dots)); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if n != len(dots) {
		t.Errorf("Invalid write count: %d", n)
	}

	// Write abc
	if n, err := f.Write([]byte(abc)); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if n != len(abc) {
		t.Errorf("Invalid write count: %d", n)
	}

	// Seek to beginning of file
	if n, err := f.Seek(0, os.SEEK_SET); err != nil || n != 0 {
		t.Errorf("Seek error: %d %s", n, err)
	}

	// Seek to end of file
	if n, err := f.Seek(0, os.SEEK_END); err != nil || n != 32 {
		t.Errorf("Seek error: %d %s", n, err)
	}

	// Write dots at end of file
	if n, err := f.Write([]byte(dots)); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if n != len(dots) {
		t.Errorf("Invalid write count: %d", n)
	}

	// Seek to beginning of file
	if n, err := f.Seek(0, os.SEEK_SET); err != nil || n != 0 {
		t.Errorf("Seek error: %d %s", n, err)
	}

	p := make([]byte, len(dots)+len(abc)+len(dots))
	if n, err := f.Read(p); err != nil || n != len(dots)+len(abc)+len(dots) {
		t.Errorf("Read error: %d %s", n, err)
	} else if s := string(p); s != dots+abc+dots {
		t.Errorf("Invalid read: %s", s)
	}
}

func TestOpenRO(t *testing.T) {
	fs := Create()
	f, err := fs.OpenFile("/readme.txt", os.O_CREATE|os.O_RDONLY, 0666)
	if err != nil {
		t.Fatalf("Could not open file: %s", err)
	}

	// Write first dots
	if _, err := f.Write([]byte(dots)); err == nil {
		t.Fatalf("Expected write error")
	}
	f.Close()
}

func TestOpenWO(t *testing.T) {
	fs := Create()
	f, err := fs.OpenFile("/readme.txt", os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		t.Fatalf("Could not open file: %s", err)
	}

	// Write first dots
	if n, err := f.Write([]byte(dots)); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if n != len(dots) {
		t.Errorf("Invalid write count: %d", n)
	}

	// Seek to beginning of file
	if n, err := f.Seek(0, os.SEEK_SET); err != nil || n != 0 {
		t.Errorf("Seek error: %d %s", n, err)
	}

	// Try reading
	p := make([]byte, len(dots))
	if n, err := f.Read(p); err == nil || n > 0 {
		t.Errorf("Expected invalid read: %d %s", n, err)
	}

	f.Close()
}

func TestOpenAppend(t *testing.T) {
	fs := Create()
	f, err := fs.OpenFile("/readme.txt", os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		t.Fatalf("Could not open file: %s", err)
	}

	// Write first dots
	if n, err := f.Write([]byte(dots)); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if n != len(dots) {
		t.Errorf("Invalid write count: %d", n)
	}
	f.Close()

	// Reopen file in append mode
	f, err = fs.OpenFile("/readme.txt", os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		t.Fatalf("Could not open file: %s", err)
	}

	// append dots
	if n, err := f.Write([]byte(abc)); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if n != len(abc) {
		t.Errorf("Invalid write count: %d", n)
	}

	// Seek to beginning of file
	if n, err := f.Seek(0, os.SEEK_SET); err != nil || n != 0 {
		t.Errorf("Seek error: %d %s", n, err)
	}

	p := make([]byte, len(dots)+len(abc))
	if n, err := f.Read(p); err != nil || n != len(dots)+len(abc) {
		t.Errorf("Read error: %d %s", n, err)
	} else if s := string(p); s != dots+abc {
		t.Errorf("Invalid read: %s", s)
	}
	f.Close()
}

func TestTruncateToLength(t *testing.T) {
	var params = []struct {
		size int64
		err  bool
	}{
		{-1, true},
		{0, false},
		{int64(len(dots) - 1), false},
		{int64(len(dots)), false},
		{int64(len(dots) + 1), false},
	}
	for _, param := range params {
		fs := Create()
		f, err := fs.OpenFile("/readme.txt", os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			t.Fatalf("Could not open file: %s", err)
		}
		if n, err := f.Write([]byte(dots)); err != nil {
			t.Errorf("Unexpected error: %s", err)
		} else if n != len(dots) {
			t.Errorf("Invalid write count: %d", n)
		}
		f.Close()

		newSize := param.size
		err = f.Truncate(newSize)
		if param.err {
			if err == nil {
				t.Errorf("Error expected truncating file to length %d", newSize)
			}
			return
		} else if err != nil {
			t.Errorf("Error truncating file: %s", err)
		}

		b, err := readFile(fs, "/readme.txt")
		if err != nil {
			t.Errorf("Error reading truncated file: %s", err)
		}
		if int64(len(b)) != newSize {
			t.Errorf("File should be empty after truncation: %d", len(b))
		}
		if fi, err := fs.Stat("/readme.txt"); err != nil {
			t.Errorf("Error stat file: %s", err)
		} else if fi.Size() != newSize {
			t.Errorf("Filesize should be %d after truncation", newSize)
		}
	}
}

func TestTruncateToZero(t *testing.T) {
	const content = "read me"
	fs := Create()
	if _, err := writeFile(fs, "/readme.txt", os.O_CREATE|os.O_RDWR, 0666, []byte(content)); err != nil {
		t.Errorf("Unexpected error writing file: %s", err)
	}

	f, err := fs.OpenFile("/readme.txt", os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		t.Errorf("Error opening file truncated: %s", err)
	}
	f.Close()

	b, err := readFile(fs, "/readme.txt")
	if err != nil {
		t.Errorf("Error reading truncated file: %s", err)
	}
	if len(b) != 0 {
		t.Errorf("File should be empty after truncation")
	}
	if fi, err := fs.Stat("/readme.txt"); err != nil {
		t.Errorf("Error stat file: %s", err)
	} else if fi.Size() != 0 {
		t.Errorf("Filesize should be 0 after truncation")
	}
}

func TestStat(t *testing.T) {
	fs := Create()
	f, err := fs.OpenFile("/readme.txt", os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		t.Fatalf("Could not open file: %s", err)
	}

	// Write first dots
	if n, err := f.Write([]byte(dots)); err != nil {
		t.Fatalf("Unexpected error: %s", err)
	} else if n != len(dots) {
		t.Fatalf("Invalid write count: %d", n)
	}
	f.Close()

	if err := fs.Mkdir("/tmp", 0777); err != nil {
		t.Fatalf("Mkdir error: %s", err)
	}

	fi, err := fs.Stat(f.Name())
	if err != nil {
		t.Errorf("Stat error: %s", err)
	}

	// Fileinfo name is base name
	if name := fi.Name(); name != "readme.txt" {
		t.Errorf("Invalid fileinfo name: %s", name)
	}

	// File name is abs name
	if name := f.Name(); name != "/readme.txt" {
		t.Errorf("Invalid file name: %s", name)
	}

	if s := fi.Size(); s != int64(len(dots)) {
		t.Errorf("Invalid size: %d", s)
	}
	if fi.IsDir() {
		t.Errorf("Invalid IsDir")
	}
	if m := fi.Mode(); m != 0666 {
		t.Errorf("Invalid mode: %d", m)
	}
}

func writeFile(fs vfs.Filesystem, name string, flags int, mode os.FileMode, b []byte) (int, error) {
	f, err := fs.OpenFile(name, flags, mode)
	if err != nil {
		return 0, err
	}
	return f.Write(b)
}

func readFile(fs vfs.Filesystem, name string) ([]byte, error) {
	f, err := fs.OpenFile(name, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(f)
}

func TestRename(t *testing.T) {
	const content = "read me"
	fs := Create()
	if _, err := writeFile(fs, "/readme.txt", os.O_CREATE|os.O_RDWR, 0666, []byte(content)); err != nil {
		t.Errorf("Unexpected error writing file: %s", err)
	}

	if err := fs.Rename("/readme.txt", "/README.txt"); err != nil {
		t.Errorf("Unexpected error renaming file: %s", err)
	}

	if _, err := fs.Stat("/readme.txt"); err == nil {
		t.Errorf("Old file still exists")
	}

	if _, err := fs.Stat("/README.txt"); err != nil {
		t.Errorf("Error stat newfile: %s", err)
	}
	if b, err := readFile(fs, "/README.txt"); err != nil {
		t.Errorf("Error reading file: %s", err)
	} else if s := string(b); s != content {
		t.Errorf("Invalid content: %s", s)
	}

	// Rename unknown file
	if err := fs.Rename("/nonexisting.txt", "/goodtarget.txt"); err == nil {
		t.Errorf("Expected error renaming file")
	}

	// Rename unknown file in nonexisting directory
	if err := fs.Rename("/nonexisting/nonexisting.txt", "/goodtarget.txt"); err == nil {
		t.Errorf("Expected error renaming file")
	}

	// Rename existing file to nonexisting directory
	if err := fs.Rename("/README.txt", "/nonexisting/nonexisting.txt"); err == nil {
		t.Errorf("Expected error renaming file")
	}

	if err := fs.Mkdir("/newdirectory", 0777); err != nil {
		t.Errorf("Error creating directory: %s", err)
	}

	if err := fs.Rename("/README.txt", "/newdirectory/README.txt"); err != nil {
		t.Errorf("Error renaming file: %s", err)
	}

	// Create the same file again at root
	if _, err := writeFile(fs, "/README.txt", os.O_CREATE|os.O_RDWR, 0666, []byte(content)); err != nil {
		t.Errorf("Unexpected error writing file: %s", err)
	}

	// Overwrite existing file
	if err := fs.Rename("/newdirectory/README.txt", "/README.txt"); err == nil {
		t.Errorf("Expected error renaming file")
	}

}

func TestModTime(t *testing.T) {
	fs := Create()

	tBeforeWrite := time.Now()
	writeFile(fs, "/readme.txt", os.O_CREATE|os.O_RDWR, 0666, []byte{0, 0, 0})
	fi, _ := fs.Stat("/readme.txt")
	mtimeAfterWrite := fi.ModTime()

	if !mtimeAfterWrite.After(tBeforeWrite) {
		t.Error("Open should modify mtime")
	}

	f, err := fs.OpenFile("/readme.txt", os.O_RDONLY, 0666)
	if err != nil {
		t.Fatalf("Could not open file: %s", err)
	}
	f.Close()
	tAfterRead := fi.ModTime()

	if tAfterRead != mtimeAfterWrite {
		t.Error("Open with O_RDONLY should not modify mtime")
	}
}
