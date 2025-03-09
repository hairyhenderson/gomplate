package funcs

import (
	"context"
	"io/fs"
	"path/filepath"

	osfs "github.com/hack-pad/hackpadfs/os"
	"github.com/hairyhenderson/gomplate/v4/conv"
	"github.com/hairyhenderson/gomplate/v4/internal/datafs"
	"github.com/hairyhenderson/gomplate/v4/internal/iohelpers"
)

// CreateFileFuncs -
func CreateFileFuncs(ctx context.Context) map[string]any {
	fsys, err := datafs.FSysForPath(ctx, "/")
	if err != nil {
		fsys = datafs.WrapWdFS(osfs.NewFS())
	}

	ns := &FileFuncs{
		ctx: ctx,
		fs:  fsys,
	}

	return map[string]any{
		"file": func() any { return ns },
	}
}

// FileFuncs -
type FileFuncs struct {
	ctx context.Context
	fs  fs.FS
}

// Read -
func (f *FileFuncs) Read(path any) (string, error) {
	b, err := fs.ReadFile(f.fs, conv.ToString(path))
	return string(b), err
}

// Stat -
func (f *FileFuncs) Stat(path any) (fs.FileInfo, error) {
	return fs.Stat(f.fs, conv.ToString(path))
}

// Exists -
func (f *FileFuncs) Exists(path any) bool {
	_, err := f.Stat(conv.ToString(path))
	return err == nil
}

// IsDir -
func (f *FileFuncs) IsDir(path any) bool {
	i, err := f.Stat(conv.ToString(path))
	return err == nil && i.IsDir()
}

// ReadDir -
func (f *FileFuncs) ReadDir(path any) ([]string, error) {
	des, err := fs.ReadDir(f.fs, conv.ToString(path))
	if err != nil {
		return nil, err
	}

	names := make([]string, len(des))
	for i, de := range des {
		names[i] = de.Name()
	}

	return names, nil
}

// Walk -
func (f *FileFuncs) Walk(path any) ([]string, error) {
	files := make([]string, 0)
	err := fs.WalkDir(f.fs, conv.ToString(path), func(subpath string, _ fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// fs.WalkDir always uses slash-separated paths, even on Windows. We
		// need to convert them to the OS-specific separator as that was the
		// previous behavior.
		subpath = filepath.FromSlash(subpath)

		files = append(files, subpath)
		return nil
	})
	return files, err
}

// Write -
func (f *FileFuncs) Write(path any, data any) (s string, err error) {
	type byteser interface{ Bytes() []byte }

	var content []byte
	fname := conv.ToString(path)

	if b, ok := data.([]byte); ok {
		content = b
	} else if b, ok := data.(byteser); ok {
		content = b.Bytes()
	} else {
		content = []byte(conv.ToString(data))
	}

	err = iohelpers.WriteFile(f.fs, fname, content)

	return "", err
}
