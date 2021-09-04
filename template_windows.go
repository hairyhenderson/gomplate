//go:build windows
// +build windows

package gomplate

import (
	"os"

	"golang.org/x/sys/windows"
)

func isDirError(name string) *os.PathError {
	return &os.PathError{
		Op:   "open",
		Path: name,
		Err:  windows.ERROR_INVALID_HANDLE,
	}
}
