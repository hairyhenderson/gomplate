//go:build !windows
// +build !windows

package gomplate

import (
	"os"

	"golang.org/x/sys/unix"
)

func isDirError(name string) *os.PathError {
	return &os.PathError{
		Op:   "open",
		Path: name,
		Err:  unix.EISDIR,
	}
}
