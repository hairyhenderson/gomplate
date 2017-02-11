package mountfs

import (
	"errors"
	"github.com/blang/vfs"
	"os"
	"testing"
)

type mountTest struct {
	mounts  []string
	results map[string]mtres
}

type mtres struct {
	mountPath string
	innerPath string
}

var mountTests = []mountTest{
	{
		[]string{
			"/tmp",
		},
		map[string]mtres{
			"/":         {"/", "/"},
			"/tmp":      {"/tmp", "/"},
			"/tmp/test": {"/tmp", "/test"},
		},
	},

	{
		[]string{
			"/home",
			"/home/user1",
			"/home/user2",
		},
		map[string]mtres{
			"/":                            {"/", "/"},
			"/tmp":                         {"/", "/tmp"},
			"/tmp/test":                    {"/", "/tmp/test"},
			"/home":                        {"/home", "/"},
			"/home/user1":                  {"/home/user1", "/"},
			"/home/user2":                  {"/home/user2", "/"},
			"/home/user1/subdir":           {"/home/user1", "/subdir"},
			"/home/user2/subdir/subsubdir": {"/home/user2", "/subdir/subsubdir"},
		},
	},
}

func TestInterface(t *testing.T) {
	_ = vfs.Filesystem(Create(nil))
}

func TestFindMount(t *testing.T) {
	for i, mtest := range mountTests {
		mounts := make(map[string]vfs.Filesystem)
		revmounts := make(map[vfs.Filesystem]string)
		for _, mount := range mtest.mounts {
			fs := vfs.Dummy(nil)
			mounts[mount] = fs
			revmounts[fs] = mount
		}
		fallback := vfs.Dummy(nil)
		for path, expRes := range mtest.results {
			expMountPath := expRes.mountPath
			expInnerPath := expRes.innerPath

			res, resInnerPath := findMount(path, mounts, fallback, "/")
			if res == nil {
				t.Errorf("Got nil")
				continue
			}
			if res == fallback {
				if expMountPath != "/" {
					t.Fatalf("Invalid mount result test case %d, mounts: %q, path: %q, expected: %q, got: %q", i, mtest.mounts, path, expMountPath, "/")
				}
				continue
			}

			if resMountPath, ok := revmounts[res]; ok {
				if resMountPath != expMountPath {
					t.Fatalf("Invalid mount, test case %d, mounts: %q, path: %q, expected: %q, got: %q", i, mtest.mounts, path, expMountPath, resMountPath)
				}
				if resInnerPath != expInnerPath {
					t.Fatalf("Invalid inner path, test case %d, mounts: %q, path: %q, expected: %q, got: %q", i, mtest.mounts, path, expInnerPath, resInnerPath)
				}
				continue
			}
			t.Fatalf("Invalid mount result test case %d, mounts: %q, path: %q, expected: %q, got invalid result", i, mtest.mounts, path, expMountPath)
		}
	}
}

type testDummyFS struct {
	vfs.Filesystem
	lastPath  string
	lastPath2 string
}

func (fs testDummyFS) OpenFile(name string, flag int, perm os.FileMode) (vfs.File, error) {
	return vfs.DummyFile(errors.New("Mount")), nil
}

func TestCreate(t *testing.T) {
	errRoot := errors.New("Rootfs")
	errMount := errors.New("Mount")
	rootFS := vfs.Dummy(errRoot)
	mountFS := vfs.Dummy(errMount)
	fs := Create(rootFS)
	if err := fs.Mkdir("/dir", 0); err != errRoot {
		t.Errorf("Expected error from rootFS: %s", err)
	}

	fs.Mount(mountFS, "/tmp")
	if err := fs.Mkdir("/tmp/dir", 0); err != errMount {
		t.Errorf("Expected error from mountFS: %s", err)
	}

	// Change rootfs
	fs.Mount(mountFS, "/")
	if err := fs.Mkdir("/dir2", 0); err != errMount {
		t.Errorf("Expected error from mountFS: %s", err)
	}
}

func TestOpenFile(t *testing.T) {
	errRoot := errors.New("Rootfs")
	errMount := errors.New("Mount")
	rootFS := vfs.Dummy(errRoot)
	mountFS := &testDummyFS{Filesystem: vfs.Dummy(errMount)}
	fs := Create(rootFS)
	fs.Mount(mountFS, "/tmp")

	// Test selection of correct fs
	f, err := fs.OpenFile("/tmp/testfile", os.O_CREATE, 0)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if n := f.Name(); n != "/tmp/testfile" {
		t.Errorf("Unexpected filename: %s", n)
	}
}

func (fs *testDummyFS) Mkdir(name string, perm os.FileMode) error {
	fs.lastPath = name
	return fs.Filesystem.Mkdir(name, perm)
}

func TestMkdir(t *testing.T) {
	errRoot := errors.New("Rootfs")
	errMount := errors.New("Mount")
	rootFS := vfs.Dummy(errRoot)
	mountFS := &testDummyFS{Filesystem: vfs.Dummy(errMount)}
	fs := Create(rootFS)
	fs.Mount(mountFS, "/tmp")

	// Test selection of correct fs
	err := fs.Mkdir("/tmp/testdir", 0)
	if err != errMount {
		t.Errorf("Wrong filesystem selected: %s", err)
	}
	if n := mountFS.lastPath; n != "/testdir" {
		t.Errorf("Incorrect inner name: %s", n)
	}
}

func (fs *testDummyFS) Remove(name string) error {
	fs.lastPath = name
	return fs.Filesystem.Remove(name)
}

func TestRemove(t *testing.T) {
	errRoot := errors.New("Rootfs")
	errMount := errors.New("Mount")
	rootFS := vfs.Dummy(errRoot)
	mountFS := &testDummyFS{Filesystem: vfs.Dummy(errMount)}
	fs := Create(rootFS)
	fs.Mount(mountFS, "/tmp")

	// Test selection of correct fs
	err := fs.Remove("/tmp/testdir")
	if err != errMount {
		t.Errorf("Wrong filesystem selected: %s", err)
	}
	if n := mountFS.lastPath; n != "/testdir" {
		t.Errorf("Incorrect inner name: %s", n)
	}
}

func (fs *testDummyFS) Rename(oldpath, newpath string) error {
	fs.lastPath = oldpath
	fs.lastPath2 = newpath
	return fs.Filesystem.Rename(oldpath, newpath)
}

func TestRename(t *testing.T) {
	errRoot := errors.New("Rootfs")
	errMount := errors.New("Mount")
	rootFS := vfs.Dummy(errRoot)
	mountFS := &testDummyFS{Filesystem: vfs.Dummy(errMount)}
	fs := Create(rootFS)
	fs.Mount(mountFS, "/tmp")

	// Test selection of correct fs
	err := fs.Rename("/tmp/testfile1", "/tmp/testfile2")
	if err != errMount {
		t.Errorf("Wrong filesystem selected: %s", err)
	}
	if n := mountFS.lastPath; n != "/testfile1" {
		t.Errorf("Incorrect inner name (oldpath): %s", n)
	}
	if n := mountFS.lastPath2; n != "/testfile2" {
		t.Errorf("Incorrect inner name (newpath): %s", n)
	}
}

func TestRenameBoundaries(t *testing.T) {
	errRoot := errors.New("Rootfs")
	errMount := errors.New("Mount")
	rootFS := vfs.Dummy(errRoot)
	mountFS := vfs.Dummy(errMount)
	fs := Create(rootFS)
	fs.Mount(mountFS, "/tmp")

	// Test selection of correct fs
	err := fs.Rename("/tmp/testfile1", "/testfile2")
	if err != ErrBoundary {
		t.Errorf("Invalid error, should return boundaries error: %s", err)
	}
}

func (fs *testDummyFS) Stat(name string) (os.FileInfo, error) {
	fs.lastPath = name
	return vfs.DumFileInfo{
		IName: name,
	}, nil
}

func TestStat(t *testing.T) {
	errRoot := errors.New("Rootfs")
	errMount := errors.New("Mount")
	rootFS := vfs.Dummy(errRoot)
	mountFS := &testDummyFS{Filesystem: vfs.Dummy(errMount)}
	fs := Create(rootFS)
	fs.Mount(mountFS, "/tmp")

	// Test selection of correct fs
	_, err := fs.Stat("/tmp/testfile1")
	if err != nil {
		t.Errorf("Wrong filesystem selected: %s", err)
	}

	if n := mountFS.lastPath; n != "/testfile1" {
		t.Errorf("Incorrect inner name: %s", n)
	}
}

func TestStatMountPoint(t *testing.T) {
	errRoot := errors.New("Rootfs")
	errMount := errors.New("Mount")
	rootFS := vfs.Dummy(errRoot)
	mountFS := &testDummyFS{Filesystem: vfs.Dummy(errMount)}
	fs := Create(rootFS)
	fs.Mount(mountFS, "/tmp")

	// Test selection of correct fs
	fi, err := fs.Stat("/tmp")
	if err != nil {
		t.Errorf("Wrong filesystem selected: %s", err)
	}

	if n := mountFS.lastPath; n != "/" {
		t.Errorf("Incorrect inner name: %s", n)
	}

	if n := fi.Name(); n != "tmp" {
		t.Errorf("Mountpoint should be return correct name, got instead: %s", n)
	}
}

func (fs *testDummyFS) Lstat(name string) (os.FileInfo, error) {
	fs.lastPath = name
	return vfs.DumFileInfo{
		IName: name,
	}, nil
}

func TestLstat(t *testing.T) {
	errRoot := errors.New("Rootfs")
	errMount := errors.New("Mount")
	rootFS := vfs.Dummy(errRoot)
	mountFS := &testDummyFS{Filesystem: vfs.Dummy(errMount)}
	fs := Create(rootFS)
	fs.Mount(mountFS, "/tmp")

	// Test selection of correct fs
	_, err := fs.Lstat("/tmp/testfile1")
	if err != nil {
		t.Errorf("Wrong filesystem selected: %s", err)
	}

	if n := mountFS.lastPath; n != "/testfile1" {
		t.Errorf("Incorrect inner name: %s", n)
	}
}

func TestLstatMountPoint(t *testing.T) {
	errRoot := errors.New("Rootfs")
	errMount := errors.New("Mount")
	rootFS := vfs.Dummy(errRoot)
	mountFS := &testDummyFS{Filesystem: vfs.Dummy(errMount)}
	fs := Create(rootFS)
	fs.Mount(mountFS, "/tmp")

	// Test selection of correct fs
	fi, err := fs.Lstat("/tmp")
	if err != nil {
		t.Errorf("Wrong filesystem selected: %s", err)
	}

	if n := mountFS.lastPath; n != "/" {
		t.Errorf("Incorrect inner name: %s", n)
	}

	if n := fi.Name(); n != "tmp" {
		t.Errorf("Mountpoint should be return correct name, got instead: %s", n)
	}
}

func (fs *testDummyFS) ReadDir(path string) ([]os.FileInfo, error) {
	fs.lastPath = path
	return []os.FileInfo{
		vfs.DumFileInfo{
			IName: "testcontent",
		},
	}, nil
}

func TestReadDir(t *testing.T) {
	errRoot := errors.New("Rootfs")
	errMount := errors.New("Mount")
	rootFS := vfs.Dummy(errRoot)
	mountFS := &testDummyFS{Filesystem: vfs.Dummy(errMount)}
	fs := Create(rootFS)
	fs.Mount(mountFS, "/tmp")

	// Test selection of correct fs
	fis, err := fs.ReadDir("/tmp")
	if err != nil {
		t.Errorf("Wrong filesystem selected: %s", err)
	}

	if n := mountFS.lastPath; n != "/" {
		t.Errorf("Incorrect inner name: %s", n)
	}

	if len(fis) != 1 || fis[0].Name() != "testcontent" {
		t.Errorf("Expected fake file, but got: %s", fis)
	}
}

func TestReadDirMountPoints(t *testing.T) {
	errRoot := errors.New("Rootfs")
	errMount := errors.New("Mount")
	rootFS := &testDummyFS{Filesystem: vfs.Dummy(errRoot)}
	mountFS := &testDummyFS{Filesystem: vfs.Dummy(errMount)}
	fs := Create(rootFS)
	fs.Mount(mountFS, "/tmp")

	// Test selection of correct fs
	fis, err := fs.ReadDir("/")
	if err != nil {
		t.Errorf("Wrong filesystem selected: %s", err)
	}

	if n := rootFS.lastPath; n != "/" {
		t.Errorf("Incorrect inner name: %s", n)
	}

	if l := len(fis); l != 2 {
		t.Fatalf("Expected 2 files, one fake, one mountpoint: %d, %s", l, fis)
	}
	if n := fis[0].Name(); n != "testcontent" {
		t.Errorf("Expected fake file, but got: %s", fis)
	}
	if n := fis[1].Name(); n != "tmp" {
		t.Errorf("Expected mountpoint, but got: %s", fis)
	}
}
