package datafs

import (
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/hack-pad/hackpadfs"
	osfs "github.com/hack-pad/hackpadfs/os"
	"github.com/hairyhenderson/go-fsimpl"
)

// ResolveLocalPath resolves a path on the given filesystem, relative to the
// current working directory, and returns both the root (/ or a volume name on
// Windows) and the resolved path. If the path is absolute (e.g. starts with a `/` or
// volume name on Windows), it is split and returned as-is.
// If fsys is nil, the current working directory is used.
// The output is suitable for use with [io/fs] functions.
func ResolveLocalPath(fsys fs.FS, name string) (root, resolved string, err error) {
	// ignore empty names
	if len(name) == 0 {
		return "", "", nil
	}

	switch fsys := fsys.(type) {
	case *wdFS:
		return resolveLocalPath(fsys.vol, name)
	default:
	}

	vol, err := workingVolume()
	if err != nil {
		return "", "", err
	}

	return resolveLocalPath(vol, name)
}

// workingVolume - returns the current working directory's volume name, or "/" if
// the current working directory has no volume name (e.g. on Unix).
func workingVolume() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getwd: %w", err)
	}

	vol := filepath.VolumeName(wd)
	if vol == "" {
		vol = "/"
	}

	return vol, nil
}

func resolveLocalPath(wvol, name string) (root, resolved string, err error) {
	// ignore empty names
	if len(name) == 0 {
		return "", "", nil
	}

	// we want to assume a / separator regardless of the OS
	name = filepath.ToSlash(name)

	// special-case for (Windows) paths that start with '/' but have no volume
	// name (e.g. "/foo/bar"). UNC paths (beginning with "//") are ignored.
	if name[0] == '/' && (len(name) == 1 || (name[1] != '/' && name[1] != '?')) {
		name = filepath.Join(wvol, name)
	} else if !filepath.IsAbs(name) {
		wd, _ := os.Getwd()
		name = filepath.Join(wd, name)
	}

	name, err = normalizeWindowsPath(name)
	if err != nil {
		return "", "", fmt.Errorf("normalize %q: %w", name, err)
	}

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

	return root, name, nil
}

// normalizeWindowsPath - converts the various types of Windows paths to either
// a rooted or relative path, depending on the type of path.
func normalizeWindowsPath(name string) (string, error) {
	name = strings.ReplaceAll(name, `\`, "/")

	switch win32PathType(name) {
	case winPathUnknown, winPathRootLocalDevice, winPathDriveRelative, winPathNT:
		return "", fmt.Errorf("unsupported path %q: %w", name, fs.ErrInvalid)
	case winPathDriveAbsolute, winPathRelative, winPathRooted:
		// absolute/relative returned as-is
		return name, nil
	case winPathUncAbsolute:
		// UNC paths are returned as-is
		return name, nil
	case winPathLocalDevice:
		// local device paths have the prefix stripped
		return name[4:], nil
	default:
		return "", fmt.Errorf("unknown path type %q: %w", name, fs.ErrInvalid)
	}
}

type winPathtype int

// There are 8 types of "DOS" paths in Windows (as opposed to NT paths):
//
// NT paths begin with a "\??\" prefix, and are implicitly absolute.
const (
	// - Unknown - e.g. "" or some other invalid path
	winPathUnknown winPathtype = iota
	// - Drive Absolute - e.g. C:\foo\bar
	winPathDriveAbsolute
	// - Drive Relative - e.g. C:foo\bar
	winPathDriveRelative
	// - Rooted - e.g. \foo\bar
	winPathRooted
	// - Relative - e.g. foo\bar, .\foo\bar, ..\foo\bar
	winPathRelative
	// - UNC Absolute - e.g. \\foo\bar
	winPathUncAbsolute
	// - Local Device - e.g. \\.\C:\foo\bar, \\.\COM1, \\?\C:\foo\bar
	winPathLocalDevice
	// - Root Local Device - e.g. \\. or \\?
	winPathRootLocalDevice
	// - NT path - e.g. \??\C:\foo\bar or \??\UNC\foo\bar
	winPathNT
)

// win32PathType - returns the type of path, as defined by the win32Path enum
// See https://googleprojectzero.blogspot.com/2016/02/the-definitive-guide-on-win32-to-nt.html
// for details on the different types
func win32PathType(name string) winPathtype {
	if name == "" {
		return winPathUnknown
	}

	// not using filepath.ToSlash here, because we want to be able to test this
	// on non-Windows systems too
	name = strings.ReplaceAll(name, `\`, "/")

	// if the first character is a slash, it's either rooted, a UNC, a local device, or root local device path
	if name[0] == '/' {
		switch {
		case len(name) == 1 || (name[1] != '/' && name[1] != '?'):
			return winPathRooted
		case len(name) == 2 || (name[2] != '.' && name[2] != '?'):
			return winPathUncAbsolute
		case len(name) >= 4 && name[1:4] == "??/":
			return winPathNT
		case len(name) >= 4 && name[3] == '/':
			return winPathLocalDevice
		default:
			return winPathRootLocalDevice
		}
	}

	switch {
	case len(name) == 1 || name[1] != ':':
		return winPathRelative
	case len(name) == 2 || name[2] != '/':
		return winPathDriveRelative
	default:
		return winPathDriveAbsolute
	}
}

func isSupportedPath(name string) bool {
	switch win32PathType(name) {
	case winPathUnknown, winPathRootLocalDevice, winPathDriveRelative, winPathNT:
		return false
	default:
		return true
	}
}

// WdFS is a filesystem provider that creates local filesystems which support
// absolute paths beginning with '/', and interpret relative paths as relative
// to the current working directory (as reported by [os.Getwd]).
//
// On Windows, certain types of paths are not supported, and will return an
// error. These are:
// - Drive Relative - e.g. C:foo\bar
// - Root Local - e.g. \\. or \\?
// - non-drive Local Devices - e.g. \\.\COM1, \\.\pipe\foo
// - NT Paths - e.g. \??\C:\foo\bar or \??\UNC\foo\bar
var WdFS = fsimpl.FSProviderFunc(
	func(u *url.URL) (fs.FS, error) {
		if !isSupportedPath(u.Path) {
			return nil, fmt.Errorf("unsupported path %q: %w", u.Path, fs.ErrInvalid)
		}

		vol, _, err := ResolveLocalPath(nil, u.Path)
		if err != nil {
			return nil, fmt.Errorf("resolve %q: %w", u.Path, err)
		}

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

	vol, _ := workingVolume()

	return &wdFS{vol: vol, fsys: fsys}
}

// wdFS is a filesystem wrapper that assumes non-absolute paths are relative to
// the current working directory (as reported by [os.Getwd]).
// It only works in a meaningful way when used with a local filesystem (e.g.
// [os.DirFS] or [hackpadfs/os.FS]).
type wdFS struct {
	fsys fs.FS

	// volume name used for drive-relative paths on Windows for cases when they
	// shouldn't be relative to the current working directory's volume
	// TODO: validate that this is actually needed
	vol string
}

var (
	_ fs.FS                = (*wdFS)(nil)
	_ fs.StatFS            = (*wdFS)(nil)
	_ fs.ReadFileFS        = (*wdFS)(nil)
	_ fs.ReadDirFS         = (*wdFS)(nil)
	_ fs.SubFS             = (*wdFS)(nil)
	_ fs.GlobFS            = (*wdFS)(nil)
	_ hackpadfs.CreateFS   = (*wdFS)(nil)
	_ hackpadfs.OpenFileFS = (*wdFS)(nil)
	_ hackpadfs.MkdirFS    = (*wdFS)(nil)
	_ hackpadfs.MkdirAllFS = (*wdFS)(nil)
	_ hackpadfs.RemoveFS   = (*wdFS)(nil)
	_ hackpadfs.ChmodFS    = (*wdFS)(nil)
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
	root, resolved, err := resolveLocalPath(w.vol, name)
	if err != nil {
		return nil, fmt.Errorf("resolve: %w", err)
	}
	fsys, err := w.fsysFor(root)
	if err != nil {
		return nil, err
	}
	return fsys.Open(resolved)
}

func (w *wdFS) Stat(name string) (fs.FileInfo, error) {
	root, resolved, err := resolveLocalPath(w.vol, name)
	if err != nil {
		return nil, fmt.Errorf("resolve: %w", err)
	}
	fsys, err := w.fsysFor(root)
	if err != nil {
		return nil, err
	}
	return fs.Stat(fsys, resolved)
}

func (w *wdFS) ReadFile(name string) ([]byte, error) {
	root, resolved, err := resolveLocalPath(w.vol, name)
	if err != nil {
		return nil, fmt.Errorf("resolve: %w", err)
	}
	fsys, err := w.fsysFor(root)
	if err != nil {
		return nil, err
	}
	return fs.ReadFile(fsys, resolved)
}

func (w *wdFS) ReadDir(name string) ([]fs.DirEntry, error) {
	root, resolved, err := resolveLocalPath(w.vol, name)
	if err != nil {
		return nil, fmt.Errorf("resolve: %w", err)
	}
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
	root, resolved, err := resolveLocalPath(w.vol, name)
	if err != nil {
		return nil, fmt.Errorf("resolve: %w", err)
	}
	fsys, err := w.fsysFor(root)
	if err != nil {
		return nil, err
	}
	return hackpadfs.Create(fsys, resolved)
}

func (w *wdFS) OpenFile(name string, flag int, perm fs.FileMode) (fs.File, error) {
	root, resolved, err := resolveLocalPath(w.vol, name)
	if err != nil {
		return nil, fmt.Errorf("resolve: %w", err)
	}
	fsys, err := w.fsysFor(root)
	if err != nil {
		return nil, err
	}
	return hackpadfs.OpenFile(fsys, resolved, flag, perm)
}

func (w *wdFS) Mkdir(name string, perm fs.FileMode) error {
	root, resolved, err := resolveLocalPath(w.vol, name)
	if err != nil {
		return fmt.Errorf("resolve: %w", err)
	}
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
	root, resolved, err := resolveLocalPath(w.vol, name)
	if err != nil {
		return fmt.Errorf("resolve: %w", err)
	}
	fsys, err := w.fsysFor(root)
	if err != nil {
		return err
	}
	return hackpadfs.MkdirAll(fsys, resolved, perm)
}

func (w *wdFS) Remove(name string) error {
	root, resolved, err := resolveLocalPath(w.vol, name)
	if err != nil {
		return fmt.Errorf("resolve: %w", err)
	}
	fsys, err := w.fsysFor(root)
	if err != nil {
		return err
	}
	return hackpadfs.Remove(fsys, resolved)
}

func (w *wdFS) Chmod(name string, mode fs.FileMode) error {
	root, resolved, err := resolveLocalPath(w.vol, name)
	if err != nil {
		return fmt.Errorf("resolve: %w", err)
	}
	fsys, err := w.fsysFor(root)
	if err != nil {
		return err
	}
	return hackpadfs.Chmod(fsys, resolved, mode)
}
