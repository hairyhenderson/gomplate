package funcs

import (
	"os"
	"sync"

	"github.com/hairyhenderson/gomplate/conv"
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
func (f *FileFuncs) Read(path interface{}) (string, error) {
	return file.Read(conv.ToString(path))
}

// Stat -
func (f *FileFuncs) Stat(path interface{}) (os.FileInfo, error) {
	return f.fs.Stat(conv.ToString(path))
}

// Exists -
func (f *FileFuncs) Exists(path interface{}) bool {
	_, err := f.Stat(conv.ToString(path))
	return err == nil
}

// IsDir -
func (f *FileFuncs) IsDir(path interface{}) bool {
	i, err := f.Stat(conv.ToString(path))
	return err == nil && i.IsDir()
}

// ReadDir -
func (f *FileFuncs) ReadDir(path interface{}) ([]string, error) {
	return file.ReadDir(conv.ToString(path))
}

// Walk -
func (f *FileFuncs) Walk(path interface{}) ([]string, error) {
	files := make([]string, 0)
	afero.Walk(f.fs, conv.ToString(path), func(subpath string, finfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		files = append(files, subpath)
		return nil
	})
	return files, nil
}
