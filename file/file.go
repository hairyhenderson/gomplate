package file

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/afero"
)

// for overriding in tests
var fs = afero.NewOsFs()

// Read -
func Read(filename string) (string, error) {
	inFile, err := fs.OpenFile(filename, os.O_RDONLY, 0)
	if err != nil {
		return "", fmt.Errorf("failed to open %s\n%v", filename, err)
	}
	// nolint: errcheck
	defer inFile.Close()
	bytes, err := ioutil.ReadAll(inFile)
	if err != nil {
		err = fmt.Errorf("read failed for %s\n%v", filename, err)
		return "", err
	}
	return string(bytes), nil
}

// ReadDir -
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
