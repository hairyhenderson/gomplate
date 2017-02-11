package memfs_test

import (
	"github.com/blang/vfs/memfs"
)

func ExampleMemFS() {
	// Create a fully writable filesystem in memory
	fs := memfs.Create()
	// Like every other vfs.Filesytem, it could be wrapped, e.g. read-only:
	// fs = vfs.ReadOnly(fs)

	// The memory fs is completely empty, permissions are supported (e.g. Stat()) but have no effect.
	fs.Mkdir("/tmp", 0777)
}
