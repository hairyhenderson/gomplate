package funcs

import (
	"context"
	"path/filepath"

	"github.com/hairyhenderson/gomplate/v4/conv"
)

// CreateFilePathFuncs -
func CreateFilePathFuncs(ctx context.Context) map[string]interface{} {
	ns := &FilePathFuncs{ctx}

	return map[string]interface{}{
		"filepath": func() interface{} { return ns },
	}
}

// FilePathFuncs -
type FilePathFuncs struct {
	ctx context.Context
}

// Base -
func (f *FilePathFuncs) Base(in interface{}) string {
	return filepath.Base(conv.ToString(in))
}

// Clean -
func (f *FilePathFuncs) Clean(in interface{}) string {
	return filepath.Clean(conv.ToString(in))
}

// Dir -
func (f *FilePathFuncs) Dir(in interface{}) string {
	return filepath.Dir(conv.ToString(in))
}

// Ext -
func (f *FilePathFuncs) Ext(in interface{}) string {
	return filepath.Ext(conv.ToString(in))
}

// FromSlash -
func (f *FilePathFuncs) FromSlash(in interface{}) string {
	return filepath.FromSlash(conv.ToString(in))
}

// IsAbs -
func (f *FilePathFuncs) IsAbs(in interface{}) bool {
	return filepath.IsAbs(conv.ToString(in))
}

// Join -
func (f *FilePathFuncs) Join(elem ...interface{}) string {
	s := conv.ToStrings(elem...)
	return filepath.Join(s...)
}

// Match -
func (f *FilePathFuncs) Match(pattern, name interface{}) (matched bool, err error) {
	return filepath.Match(conv.ToString(pattern), conv.ToString(name))
}

// Rel -
func (f *FilePathFuncs) Rel(basepath, targpath interface{}) (string, error) {
	return filepath.Rel(conv.ToString(basepath), conv.ToString(targpath))
}

// Split -
func (f *FilePathFuncs) Split(in interface{}) []string {
	dir, file := filepath.Split(conv.ToString(in))
	return []string{dir, file}
}

// ToSlash -
func (f *FilePathFuncs) ToSlash(in interface{}) string {
	return filepath.ToSlash(conv.ToString(in))
}

// VolumeName -
func (f *FilePathFuncs) VolumeName(in interface{}) string {
	return filepath.VolumeName(conv.ToString(in))
}
