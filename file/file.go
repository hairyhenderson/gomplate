// Package file contains functions for working with files and directories on the local filesystem
package file

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/hairyhenderson/gomplate/v3/internal/iohelpers"
	"github.com/pkg/errors"

	"github.com/spf13/afero"
)

// for overriding in tests
var fs = afero.NewOsFs()

// Read the contents of the referenced file, as a string.
func Read(filename string) (string, error) {
	inFile, err := fs.OpenFile(filename, os.O_RDONLY, 0)
	if err != nil {
		return "", errors.Wrapf(err, "failed to open %s", filename)
	}
	// nolint: errcheck
	defer inFile.Close()
	bytes, err := ioutil.ReadAll(inFile)
	if err != nil {
		err = errors.Wrapf(err, "read failed for %s", filename)
		return "", err
	}
	return string(bytes), nil
}

// ReadDir gets a directory listing.
func ReadDir(path string) ([]string, error) {
	f, err := fs.Open(path)
	if err != nil {
		return nil, err
	}
	i, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if i.IsDir() {
		return f.Readdirnames(0)
	}
	return nil, errors.New("file is not a directory")
}

// Write the given content to the file, truncating any existing file, and
// creating the directory structure leading up to it if necessary.
func Write(filename string, content []byte) error {
	err := assertPathInWD(filename)
	if err != nil {
		return errors.Wrapf(err, "failed to open %s", filename)
	}

	fi, err := os.Stat(filename)
	if err != nil && !os.IsNotExist(err) {
		return errors.Wrapf(err, "failed to stat %s", filename)
	}
	mode := iohelpers.NormalizeFileMode(0o644)
	if fi != nil {
		mode = fi.Mode()
	}
	err = fs.MkdirAll(filepath.Dir(filename), 0o755)
	if err != nil {
		return errors.Wrapf(err, "failed to make dirs for %s", filename)
	}
	inFile, err := fs.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return errors.Wrapf(err, "failed to open %s", filename)
	}
	n, err := inFile.Write(content)
	if err != nil {
		return errors.Wrapf(err, "failed to write %s", filename)
	}
	if n != len(content) {
		return errors.Wrapf(err, "short write on %s (%d bytes)", filename, n)
	}
	return nil
}

func assertPathInWD(filename string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	f, err := filepath.Abs(filename)
	if err != nil {
		return err
	}
	r, err := filepath.Rel(wd, f)
	if err != nil {
		return err
	}
	if strings.HasPrefix(r, "..") {
		return errors.Errorf("path %s not contained by working directory %s (rel: %s)", filename, wd, r)
	}
	return nil
}
