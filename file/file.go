// Package file contains functions for working with files and directories on the local filesystem
package file

import (
	"fmt"
	"io/fs"

	"github.com/hairyhenderson/gomplate/v4/internal/datafs"
	"github.com/hairyhenderson/gomplate/v4/internal/iohelpers"

	osfs "github.com/hack-pad/hackpadfs/os"
)

// fsys for legacy functions - see deprecation notices
var fsys = datafs.WrapWdFS(osfs.NewFS())

// Read the contents of the referenced file, as a string.
//
// Deprecated: (as of 4.0.0) use [io/fs#ReadFile] instead
func Read(filename string) (string, error) {
	bytes, err := fs.ReadFile(fsys, filename)
	if err != nil {
		err = fmt.Errorf("read failed for %s: %w", filename, err)
		return "", err
	}
	return string(bytes), nil
}

// ReadDir gets a directory listing.
//
// Deprecated: (as of 4.0.0) use [io/fs#ReadDir] instead
func ReadDir(path string) ([]string, error) {
	des, err := fs.ReadDir(fsys, path)
	if err != nil {
		return nil, err
	}

	names := make([]string, len(des))
	for i, de := range des {
		names[i] = de.Name()
	}

	return names, nil
}

// Write the given content to the file, truncating any existing file, and
// creating the directory structure leading up to it if necessary.
//
// Deprecated: (as of 4.0.0) use [os#WriteFile] instead
func Write(filename string, content []byte) error {
	return iohelpers.WriteFile(fsys, filename, content)
}
