package funcs

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/hairyhenderson/gomplate/file"
	"github.com/spf13/afero"
)

var (
	ff     *FileFuncs
	ffInit sync.Once
)

// FileNS - the File namespace
func FileNS() *FileFuncs {
	ffInit.Do(func() { ff = &FileFuncs{afero.NewOsFs()} })
	return ff
}

// AddFileFuncs -
func AddFileFuncs(f map[string]interface{}) {
	f["file"] = FileNS
}

// FileFuncs -
type FileFuncs struct {
	fs afero.Fs
}

// Read -
func (f *FileFuncs) Read(path string) (string, error) {
	return file.Read(path)
}

// Stat -
func (f *FileFuncs) Stat(path string) (os.FileInfo, error) {
	return f.fs.Stat(path)
}

// Exists -
func (f *FileFuncs) Exists(path string) bool {
	_, err := f.Stat(path)
	return err == nil
}

// IsDir -
func (f *FileFuncs) IsDir(path string) bool {
	i, err := f.Stat(path)
	return err == nil && i.IsDir()
}

// ReadDir -
func (f *FileFuncs) ReadDir(path string) ([]string, error) {
	return file.ReadDir(path)
}

// Glob -
func (f *FileFuncs) Glob(dir string) ([]string, error) {
	fileList := make([]string, 0)
	e := filepath.Walk(dir, func(path string, finfo os.FileInfo, err error) error {
		if (!f.IsDir(path)) {
			fileList = append(fileList, path)
		}
		return err
	})

	if e != nil {
		panic(e)
	}

	return fileList, nil
}
