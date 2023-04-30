package datafs

import (
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"

	"github.com/hack-pad/hackpadfs"
	osfs "github.com/hack-pad/hackpadfs/os"
	"github.com/hairyhenderson/go-fsimpl"
)

// ResolveLocalPath resolves a path on the local filesystem, relative to the
// current working directory, and returns both the root (/ or a volume name on
// Windows) and the resolved path. If the path is absolute (e.g. starts with a `/` or
// volume name on Windows), it is split and returned as-is.
// The output is suitable for use with [io/fs] functions.
//
// TODO: maybe take fsys as an argument, and if it's a wdFS, use its vol instead
// of calling os.Getwd?
func ResolveLocalPath(name string) (root, resolved string) {
	// ignore empty names
	if len(name) == 0 {
		return "", ""
	}

	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	vol := filepath.VolumeName(wd)
	if vol == "" {
		vol = "/"
	}

	f := &wdFS{vol: vol}
	return f.resolveLocalPath(name)
}

func (w *wdFS) resolveLocalPath(name string) (root, resolved string) {
	// ignore empty names
	if len(name) == 0 {
		return "", ""
	}

	// we want to assume a / separator regardless of the OS
	name = filepath.ToSlash(name)

	// special-case for (Windows) paths that start with '/' but have no volume
	// name (e.g. "/foo/bar"). UNC paths (beginning with "//") are ignored.
	if name[0] == '/' && (len(name) == 1 || name[1] != '/') {
		name = filepath.Join(w.vol, name)
	} else if name[0] != '/' && !filepath.IsAbs(name) {
		wd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		name = filepath.Join(wd, name)
	}

	name = filepath.ToSlash(name)

	vol := filepath.VolumeName(name)
	if vol != "" && name != vol {
		root = vol
		name = name[len(vol)+1:]
	} else if name[0] == '/' {
		root = "/"
		name = name[1:]
	}

	// there may still be backslashes in the root
	root = filepath.ToSlash(root)

	// we might've emptied name, so return "." instead
	if name == "" {
		name = "."
	}

	return root, name
}

// WdFS is a filesystem provider that creates local filesystems which support
// absolute paths beginning with '/', and interpret relative paths as relative
// to the current working directory (as reported by [os.Getwd])
var WdFS = fsimpl.FSProviderFunc(
	func(u *url.URL) (fs.FS, error) {
		vol, _ := ResolveLocalPath(u.Path)

		var fsys fs.FS
		if vol == "" || vol == "/" {
			fsys = osfs.NewFS()
		} else {
			var err error
			fsys, err = osfs.NewFS().SubVolume(vol)
			if err != nil {
				return nil, err
			}
		}

		return &wdFS{vol: vol, fsys: fsys}, nil
	},
	// register for both file and empty scheme (empty when path is relative)
	"file", "",
)

// WrapWdFS is a filesystem wrapper that assumes non-absolute paths are relative
// to the current working directory (as reported by [os.Getwd]). It only works
// in a meaningful way when used with a local filesystem (e.g. [os.DirFS] or
// [hackpadfs/os.FS]).
func WrapWdFS(fsys fs.FS) fs.FS {
	// if fsys is a wdFS, just return it
	if _, ok := fsys.(*wdFS); ok {
		return fsys
	}

	return &wdFS{fsys: fsys}
}

// wdFS is a filesystem wrapper that assumes non-absolute paths are relative to
// the current working directory (as reported by [os.Getwd]).
// It only works in a meaningful way when used  with a local filesystem (e.g.
// [os.DirFS] or [hackpadfs/os.FS]).
type wdFS struct {
	fsys fs.FS
	vol  string
}

var (
	_ fs.FS                = &wdFS{}
	_ fs.StatFS            = &wdFS{}
	_ fs.ReadFileFS        = &wdFS{}
	_ fs.ReadDirFS         = &wdFS{}
	_ fs.SubFS             = &wdFS{}
	_ fs.GlobFS            = &wdFS{}
	_ hackpadfs.CreateFS   = &wdFS{}
	_ hackpadfs.OpenFileFS = &wdFS{}
	_ hackpadfs.MkdirFS    = &wdFS{}
	_ hackpadfs.MkdirAllFS = &wdFS{}
	_ hackpadfs.RemoveFS   = &wdFS{}
	_ hackpadfs.ChmodFS    = &wdFS{}
)

func (w *wdFS) fsysFor(vol string) (fs.FS, error) {
	if vol == "" || vol == "/" || vol == w.vol {
		return w.fsys, nil
	}

	// create a new osfs.FS here, because we can't modify the original if
	// SubVolume was already called on it.
	if _, ok := w.fsys.(*osfs.FS); ok {
		fsys, err := osfs.NewFS().SubVolume(vol)
		if err != nil {
			return nil, fmt.Errorf("fsysFor %q: %w", vol, err)
		}

		return fsys, nil
	}

	// just return the original filesystem if we're not wrapping an osfs.FS
	return w.fsys, nil
}

func (w *wdFS) Open(name string) (fs.File, error) {
	root, resolved := w.resolveLocalPath(name)
	fsys, err := w.fsysFor(root)
	if err != nil {
		return nil, err
	}
	return fsys.Open(resolved)
}

func (w *wdFS) Stat(name string) (fs.FileInfo, error) {
	root, resolved := w.resolveLocalPath(name)
	fsys, err := w.fsysFor(root)
	if err != nil {
		return nil, err
	}
	return fs.Stat(fsys, resolved)
}

func (w *wdFS) ReadFile(name string) ([]byte, error) {
	root, resolved := w.resolveLocalPath(name)
	fsys, err := w.fsysFor(root)
	if err != nil {
		return nil, err
	}
	return fs.ReadFile(fsys, resolved)
}

func (w *wdFS) ReadDir(name string) ([]fs.DirEntry, error) {
	root, resolved := w.resolveLocalPath(name)
	fsys, err := w.fsysFor(root)
	if err != nil {
		return nil, err
	}
	return fs.ReadDir(fsys, resolved)
}

func (w *wdFS) Sub(name string) (fs.FS, error) {
	// we don't resolve the name here, because this name must necessarily be
	// a path relative to the wrapped filesystem's root
	if fsys, ok := w.fsys.(fs.SubFS); ok {
		return fsys.Sub(name)
	}

	return fs.Sub(w.fsys, name)
}

func (w *wdFS) Glob(_ string) ([]string, error) {
	// I'm not sure how to handle this, so I'm just going to error for now -
	// I have no need of Glob anyway.
	return nil, fmt.Errorf("glob not supported by wdFS: %w", fs.ErrInvalid)
}

func (w *wdFS) Create(name string) (fs.File, error) {
	root, resolved := w.resolveLocalPath(name)
	fsys, err := w.fsysFor(root)
	if err != nil {
		return nil, err
	}
	return hackpadfs.Create(fsys, resolved)
}

func (w *wdFS) OpenFile(name string, flag int, perm fs.FileMode) (fs.File, error) {
	root, resolved := w.resolveLocalPath(name)
	fsys, err := w.fsysFor(root)
	if err != nil {
		return nil, err
	}
	return hackpadfs.OpenFile(fsys, resolved, flag, perm)
}

func (w *wdFS) Mkdir(name string, perm fs.FileMode) error {
	root, resolved := w.resolveLocalPath(name)
	fsys, err := w.fsysFor(root)
	if err != nil {
		return err
	}
	err = hackpadfs.Mkdir(fsys, resolved, perm)
	if err != nil {
		return fmt.Errorf("mkdir %q (resolved as %q): %w", name, resolved, err)
	}
	return nil
}

func (w *wdFS) MkdirAll(name string, perm fs.FileMode) error {
	root, resolved := w.resolveLocalPath(name)
	fsys, err := w.fsysFor(root)
	if err != nil {
		return err
	}
	return hackpadfs.MkdirAll(fsys, resolved, perm)
}

func (w *wdFS) Remove(name string) error {
	root, resolved := w.resolveLocalPath(name)
	fsys, err := w.fsysFor(root)
	if err != nil {
		return err
	}
	return hackpadfs.Remove(fsys, resolved)
}

func (w *wdFS) Chmod(name string, mode fs.FileMode) error {
	root, resolved := w.resolveLocalPath(name)
	fsys, err := w.fsysFor(root)
	if err != nil {
		return err
	}
	return hackpadfs.Chmod(fsys, resolved, mode)
}
