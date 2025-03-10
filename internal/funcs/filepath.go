package funcs

import (
	"context"
	"path/filepath"

	"github.com/hairyhenderson/gomplate/v4/conv"
)

// CreateFilePathFuncs -
func CreateFilePathFuncs(ctx context.Context) map[string]any {
	ns := &FilePathFuncs{ctx}

	return map[string]any{
		"filepath": func() any { return ns },
	}
}

// FilePathFuncs -
type FilePathFuncs struct {
	ctx context.Context
}

// Base -
func (f *FilePathFuncs) Base(in any) string {
	return filepath.Base(conv.ToString(in))
}

// Clean -
func (f *FilePathFuncs) Clean(in any) string {
	return filepath.Clean(conv.ToString(in))
}

// Dir -
func (f *FilePathFuncs) Dir(in any) string {
	return filepath.Dir(conv.ToString(in))
}

// Ext -
func (f *FilePathFuncs) Ext(in any) string {
	return filepath.Ext(conv.ToString(in))
}

// FromSlash -
func (f *FilePathFuncs) FromSlash(in any) string {
	return filepath.FromSlash(conv.ToString(in))
}

// IsAbs -
func (f *FilePathFuncs) IsAbs(in any) bool {
	return filepath.IsAbs(conv.ToString(in))
}

// Join -
func (f *FilePathFuncs) Join(elem ...any) string {
	s := conv.ToStrings(elem...)
	return filepath.Join(s...)
}

// Match -
func (f *FilePathFuncs) Match(pattern, name any) (matched bool, err error) {
	return filepath.Match(conv.ToString(pattern), conv.ToString(name))
}

// Rel -
func (f *FilePathFuncs) Rel(basepath, targpath any) (string, error) {
	return filepath.Rel(conv.ToString(basepath), conv.ToString(targpath))
}

// Split -
func (f *FilePathFuncs) Split(in any) []string {
	dir, file := filepath.Split(conv.ToString(in))
	return []string{dir, file}
}

// ToSlash -
func (f *FilePathFuncs) ToSlash(in any) string {
	return filepath.ToSlash(conv.ToString(in))
}

// VolumeName -
func (f *FilePathFuncs) VolumeName(in any) string {
	return filepath.VolumeName(conv.ToString(in))
}
