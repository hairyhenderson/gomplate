package vfs

import (
	"errors"
	"testing"
)

var errDum = errors.New("Dummy error")

func TestInterface(t *testing.T) {
	_ = Filesystem(Dummy(errDum))
}

func TestDummyFS(t *testing.T) {
	fs := Dummy(errDum)
	if _, err := fs.OpenFile("test", 0, 0); err != errDum {
		t.Errorf("OpenFile DummyError expected: %s", err)
	}
	if err := fs.Remove("test"); err != errDum {
		t.Errorf("Remove DummyError expected: %s", err)
	}
	if err := fs.Rename("old", "new"); err != errDum {
		t.Errorf("Rename DummyError expected: %s", err)
	}
	if err := fs.Mkdir("test", 0); err != errDum {
		t.Errorf("Mkdir DummyError expected: %s", err)
	}
	if _, err := fs.Stat("test"); err != errDum {
		t.Errorf("Stat DummyError expected: %s", err)
	}
	if _, err := fs.Lstat("test"); err != errDum {
		t.Errorf("Lstat DummyError expected: %s", err)
	}
	if _, err := fs.ReadDir("test"); err != errDum {
		t.Errorf("ReadDir DummyError expected: %s", err)
	}
}

func TestFileInterface(t *testing.T) {
	_ = File(DummyFile(errDum))
}

func TestDummyFile(t *testing.T) {
	f := DummyFile(errDum)
	if name := f.Name(); name != "dummy" {
		t.Errorf("Invalid name: %s", name)
	}
	if err := f.Close(); err != errDum {
		t.Errorf("Close DummyError expected: %s", err)
	}
	if _, err := f.Write([]byte("test")); err != errDum {
		t.Errorf("Write DummyError expected: %s", err)
	}
	if _, err := f.Read([]byte{}); err != errDum {
		t.Errorf("Read DummyError expected: %s", err)
	}
	if _, err := f.Seek(0, 0); err != errDum {
		t.Errorf("Seek DummyError expected: %s", err)
	}
	if err := f.Sync(); err != errDum {
		t.Errorf("Sync DummyError expected: %s", err)
	}
}
