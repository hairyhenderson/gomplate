package vfs_test

import (
	"errors"
	"fmt"
	"os"

	"github.com/blang/vfs"
)

type noNewDirs struct {
	vfs.Filesystem
}

func NoNewDirs(fs vfs.Filesystem) *noNewDirs {
	return &noNewDirs{fs}
}

// Mkdir is disabled
func (fs *noNewDirs) Mkdir(name string, perm os.FileMode) error {
	return errors.New("Mkdir disabled!")
}

func ExampleOsFS_myWrapper() {

	// Disable Mkdirs on the OS Filesystem
	var fs vfs.Filesystem = NoNewDirs(vfs.OS())

	err := fs.Mkdir("/tmp", 0777)
	if err != nil {
		fmt.Printf("Mkdir disabled!\n")
	}
}
