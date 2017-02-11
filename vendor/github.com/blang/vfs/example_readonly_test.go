package vfs_test

import (
	"fmt"
	"os"

	"github.com/blang/vfs"
)

// Every vfs.Filesystem could be easily wrapped
func ExampleRoFS() {
	// Create a readonly vfs accessing the filesystem of the underlying OS
	roFS := vfs.ReadOnly(vfs.OS())

	// Mkdir is disabled on ReadOnly vfs, will return vfs.ErrReadOnly
	// See vfs.ReadOnly for all disabled operations
	err := roFS.Mkdir("/tmp/vfs_example", 0777)
	if err != nil {
		fmt.Printf("Error creating directory: %s\n", err)
		return
	}

	// OpenFile is controlled to support read-only functionality. os.O_CREATE or os.O_APPEND will fail.
	// Flags like os.O_RDWR are supported but the returned file is protected e.g. from Write(..).
	f, err := roFS.OpenFile("/tmp/vfs_example/example.txt", os.O_RDWR, 0)
	if err != nil {
		fmt.Printf("Could not create file: %s\n", err)
		return

	}
	defer f.Close()

	// Will fail and return vfs.ErrReadOnly
	_, err = f.Write([]byte("VFS working on your filesystem"))
	if err != nil {
		fmt.Printf("Could not write file on read only filesystem: %s", err)
		return
	}
}
