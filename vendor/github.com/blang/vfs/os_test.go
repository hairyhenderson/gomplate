package vfs

import (
	"os"
	"testing"
)

func TestOSInterface(t *testing.T) {
	_ = Filesystem(OS())
}

func TestOSCreate(t *testing.T) {
	fs := OS()

	f, err := fs.OpenFile("/tmp/test123", os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		t.Errorf("Create: %s", err)
	}
	err = f.Close()
	if err != nil {
		t.Errorf("Close: %s", err)
	}
	err = fs.Remove(f.Name())
	if err != nil {
		t.Errorf("Remove: %s", err)
	}
}
