package vfs_test

import (
	"fmt"
	"github.com/blang/vfs"
	"github.com/blang/vfs/memfs"
	"github.com/blang/vfs/mountfs"
	"os"
)

func Example() {
	// Create a vfs accessing the filesystem of the underlying OS
	var osfs vfs.Filesystem = vfs.OS()
	osfs.Mkdir("/tmp", 0777)

	// Make the filesystem read-only:
	osfs = vfs.ReadOnly(osfs) // Simply wrap filesystems to change its behaviour

	// os.O_CREATE will fail and return vfs.ErrReadOnly
	// os.O_RDWR is supported but Write(..) on the file is disabled
	f, _ := osfs.OpenFile("/tmp/example.txt", os.O_RDWR, 0)

	// Return vfs.ErrReadOnly
	_, err := f.Write([]byte("Write on readonly fs?"))
	if err != nil {
		fmt.Errorf("Filesystem is read only!\n")
	}

	// Create a fully writable filesystem in memory
	mfs := memfs.Create()
	mfs.Mkdir("/root", 0777)

	// Create a vfs supporting mounts
	// The root fs is accessing the filesystem of the underlying OS
	fs := mountfs.Create(osfs)

	// Mount a memfs inside /memfs
	// /memfs may not exist
	fs.Mount(mfs, "/memfs")

	// This will create /testdir inside the memfs
	fs.Mkdir("/memfs/testdir", 0777)

	// This would create /tmp/testdir inside your OS fs
	// But the rootfs `osfs` is read-only
	fs.Mkdir("/tmp/testdir", 0777)
}
