package vfs_test

import (
	"fmt"

	"github.com/blang/vfs"
)

func ExampleOsFS() {

	// Create a vfs accessing the filesystem of the underlying OS
	osFS := vfs.OS()
	err := osFS.Mkdir("/tmp/vfs_example", 0777)
	if err != nil {
		fmt.Printf("Error creating directory: %s\n", err)
	}

	// Convenience method
	f, err := vfs.Create(osFS, "/tmp/vfs_example/example.txt")
	// f, err := osFS.OpenFile("/tmp/vfs/example.txt", os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Printf("Could not create file: %s\n", err)
	}
	defer f.Close()
	if _, err := f.Write([]byte("VFS working on your filesystem")); err != nil {
		fmt.Printf("Error writing to file: %s\n", err)
	}
}
