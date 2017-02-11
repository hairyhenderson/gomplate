package memfs

import (
	"github.com/blang/vfs"
	"testing"
)

func TestFileInterface(t *testing.T) {
	_ = vfs.File(NewMemFile("", nil, nil))
}
