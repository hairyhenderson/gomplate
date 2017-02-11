package mountfs_test

import (
	"github.com/blang/vfs"
	"github.com/blang/vfs/memfs"
	"github.com/blang/vfs/mountfs"
)

func ExampleMountFS() {
	// Create a vfs supporting mounts
	// The root fs is accessing the filesystem of the underlying OS
	fs := mountfs.Create(vfs.OS())

	// Mount a memfs inside /memfs
	// /memfs may not exist
	fs.Mount(memfs.Create(), "/memfs")

	// This will create /testdir inside the memfs
	fs.Mkdir("/memfs/testdir", 0777)

	// This will create /tmp/testdir inside your OS fs
	fs.Mkdir("/tmp/testdir", 0777)
}
