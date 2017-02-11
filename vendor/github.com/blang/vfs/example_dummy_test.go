package vfs_test

import (
	"errors"
	"fmt"
	"os"

	"github.com/blang/vfs"
)

type myFS struct {
	vfs.Filesystem // Embed the Filesystem interface and fill it with vfs.Dummy on creation
}

func MyFS() *myFS {
	return &myFS{
		vfs.Dummy(errors.New("Not implemented yet!")),
	}
}

func (fs myFS) Mkdir(name string, perm os.FileMode) error {
	// Create a directory
	// ...
	return nil
}

func ExampleDummyFS() {
	// Simply bootstrap your filesystem
	var fs vfs.Filesystem = MyFS()

	// Your mkdir implementation
	fs.Mkdir("/tmp", 0777)

	// All necessary methods like OpenFile (therefor Create) are stubbed
	// and return the dummys error
	_, err := vfs.Create(fs, "/tmp/vfs/example.txt")
	if err != nil {
		fmt.Printf("Error will be: Not implemented yet!\n")
	}

}
