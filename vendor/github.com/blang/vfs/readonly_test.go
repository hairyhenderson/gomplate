package vfs

import (
	"errors"
	"os"
	"testing"
)

var (
	errDummy = errors.New("Not implemented")
	// Complete dummy base
	baseFSDummy = Dummy(errDummy)
	ro          = ReadOnly(baseFSDummy)
)

func TestROInterface(t *testing.T) {
	_ = Filesystem(ro)
}

func TestROOpenFileFlags(t *testing.T) {
	_, err := ro.OpenFile("name", os.O_CREATE, 0666)
	if err != ErrReadOnly {
		t.Errorf("Create error expected")
	}

	_, err = ro.OpenFile("name", os.O_APPEND, 0666)
	if err != ErrReadOnly {
		t.Errorf("Append error expected")
	}

	_, err = ro.OpenFile("name", os.O_WRONLY, 0666)
	if err != ErrReadOnly {
		t.Errorf("WROnly error expected")
	}

	_, err = ro.OpenFile("name", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != ErrReadOnly {
		t.Errorf("Error expected")
	}

	// os.O_RDWR is allowed, dummy error is returned
	_, err = ro.OpenFile("name", os.O_RDWR, 0)
	if err != errDummy {
		t.Errorf("Expected dummy error")
	}
}

func TestRORemove(t *testing.T) {
	err := ro.Remove("test")
	if err != ErrReadOnly {
		t.Errorf("Remove error expected")
	}
}

func TestRORename(t *testing.T) {
	err := ro.Rename("old", "new")
	if err != ErrReadOnly {
		t.Errorf("Rename error expected")
	}
}

func TestMkDir(t *testing.T) {
	err := ro.Mkdir("test", 0777)
	if err != ErrReadOnly {
		t.Errorf("Mkdir error expected")
	}
}

type writeDummyFS struct {
	Filesystem
}

// Opens a dummyfile instead of error
func (fs writeDummyFS) OpenFile(name string, flag int, perm os.FileMode) (File, error) {
	return DummyFile(errDummy), nil
}

func TestROOpenFileWrite(t *testing.T) {
	// Dummy base with mocked OpenFile for write test
	roWriteMock := ReadOnly(writeDummyFS{Dummy(errDummy)})

	f, err := roWriteMock.OpenFile("name", os.O_RDWR, 0)
	if err != nil {
		t.Errorf("No OpenFile error expected: %s", err)
	}
	written, err := f.Write([]byte("test"))
	if err != ErrReadOnly {
		t.Errorf("Error expected: %s", err)
	}
	if written > 0 {
		t.Errorf("Written expected 0: %d", written)
	}
}
