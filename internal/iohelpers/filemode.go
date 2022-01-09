package iohelpers

import (
	"os"
	"runtime"
)

// NormalizeFileMode converts the given mode to a mode that will work as
// expected on the given OS. A no-op on non-Windows OSes, but on Windows modes
// work differently - only the owner read/write bits are honoured (i.e. the 0200
// mask).
func NormalizeFileMode(mode os.FileMode) os.FileMode {
	if runtime.GOOS == "windows" {
		return windowsFileMode(mode)
	}
	return mode
}

func windowsFileMode(mode os.FileMode) os.FileMode {
	// non-owner and execute bits are stripped on files
	if !mode.IsDir() {
		mode &^= 0o177
	}

	if mode&0o200 != 0 {
		// writeable implies read/write on Windows
		mode |= 0o666
	} else if mode&0o400 != 0 {
		mode |= 0o444
	}

	return mode
}
