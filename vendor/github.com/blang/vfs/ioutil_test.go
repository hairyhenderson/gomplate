package vfs_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/blang/vfs"
	"github.com/blang/vfs/memfs"
)

var (
	testpath = "/example.txt"
	testmode = os.FileMode(0600)
	testdata = bytes.Repeat([]byte("abcdefghijklmnopqrstuvwxyz"), 100)
)

func TestWriteFile(t *testing.T) {
	fs := memfs.Create()

	vfs.WriteFile(fs, testpath, testdata, testmode)

	info, err := fs.Stat(testpath)
	if err != nil {
		t.Fatalf("File not created")
	}
	if info.Size() != int64(len(testdata)) {
		t.Fatalf("Bad file size: %d bytes (expected %d)", info.Size(), len(testdata))
	}
	if info.Mode() != testmode {
		t.Fatalf("Bad file mode: %o (expected %o)", info.Mode(), testmode)
	}
}

func TestReadFile(t *testing.T) {
	fs := memfs.Create()

	f, _ := fs.OpenFile(testpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, testmode)
	f.Write(testdata)
	f.Close()

	data, err := vfs.ReadFile(fs, testpath)
	if err != nil {
		t.Fatalf("ReadFile failed: %s", err)
	}
	if len(data) != len(testdata) {
		t.Fatalf("Bad data length: %d bytes (expected %d)", len(data), len(testdata))
	}

	_, err = vfs.ReadFile(fs, "/doesnt-exist.txt")
	if err == nil {
		t.Fatalf("ReadFile failed: expected error")
	}
}
