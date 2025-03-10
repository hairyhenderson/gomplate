package datafs

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/hairyhenderson/go-fsimpl"
)

// newEnvFS returns a filesystem (an fs.FS) that can be used to read data from
// environment variables.
func newEnvFS(_ *url.URL) (fs.FS, error) {
	return &envFS{locfs: os.DirFS("/")}, nil
}

type envFS struct {
	locfs fs.FS
}

//nolint:gochecknoglobals
var EnvFS = fsimpl.FSProviderFunc(newEnvFS, "env")

var _ fs.FS = (*envFS)(nil)

func (f *envFS) Open(name string) (fs.File, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{
			Op:   "open",
			Path: name,
			Err:  fs.ErrInvalid,
		}
	}

	return &envFile{locfs: f.locfs, name: name}, nil
}

type envFile struct {
	locfs fs.FS
	body  io.Reader
	name  string

	dirents []fs.DirEntry
	diroff  int
}

var (
	_ fs.File        = (*envFile)(nil)
	_ fs.ReadDirFile = (*envFile)(nil)
)

// overridable env functions
var (
	lookupEnv = os.LookupEnv
	environ   = os.Environ
)

func (e *envFile) Close() error {
	e.body = nil
	return nil
}

func (e *envFile) envReader() (int, io.Reader, error) {
	v, found := lookupEnv(e.name)
	if found {
		return len(v), bytes.NewBufferString(v), nil
	}

	fname, found := lookupEnv(e.name + "_FILE")
	if found && fname != "" {
		fname = strings.TrimPrefix(fname, "/")

		b, err := fs.ReadFile(e.locfs, fname)
		if err != nil {
			return 0, nil, err
		}

		b = bytes.TrimSpace(b)

		return len(b), bytes.NewBuffer(b), nil
	}

	return 0, nil, fs.ErrNotExist
}

func (e *envFile) Stat() (fs.FileInfo, error) {
	n, _, err := e.envReader()
	if err != nil {
		return nil, err
	}

	return FileInfo(e.name, int64(n), 0o444, time.Time{}, ""), nil
}

func (e *envFile) Read(p []byte) (int, error) {
	if e.body == nil {
		_, r, err := e.envReader()
		if err != nil {
			return 0, err
		}
		e.body = r
	}

	return e.body.Read(p)
}

func (e *envFile) ReadDir(n int) ([]fs.DirEntry, error) {
	// envFS has no concept of subdirectories, but we can support a root
	// directory by listing all environment variables.
	if e.name != "." {
		return nil, fmt.Errorf("%s: not a directory", e.name)
	}

	if e.dirents == nil {
		envs := environ()
		e.dirents = make([]fs.DirEntry, 0, len(envs))
		for _, env := range envs {
			parts := strings.SplitN(env, "=", 2)
			name, value := parts[0], parts[1]

			if name == "" {
				// this might be a Windows =C: style env var, so skip it
				continue
			}

			e.dirents = append(e.dirents, FileInfoDirEntry(
				FileInfo(name, int64(len(value)), 0o444, time.Time{}, ""),
			))
		}
	}

	if n > 0 && e.diroff >= len(e.dirents) {
		return nil, io.EOF
	}

	low := e.diroff
	high := e.diroff + n

	// clamp high at the max, and ensure it's higher than low
	if high >= len(e.dirents) || high <= low {
		high = len(e.dirents)
	}

	entries := make([]fs.DirEntry, high-low)
	copy(entries, e.dirents[e.diroff:])

	e.diroff = high

	return entries, nil
}

// FileInfo/DirInfo/FileInfoDirEntry/etc are taken from go-fsimpl's internal
// package, and may be exported in the future...

// FileInfo creates a static fs.FileInfo with the given properties.
// The result is also a fs.DirEntry and can be safely cast.
func FileInfo(name string, size int64, mode fs.FileMode, modTime time.Time, contentType string) fs.FileInfo {
	return &staticFileInfo{
		name:        name,
		size:        size,
		mode:        mode,
		modTime:     modTime,
		contentType: contentType,
	}
}

// DirInfo creates a fs.FileInfo for a directory with the given name. Use
// FileInfo to set other values.
func DirInfo(name string, modTime time.Time) fs.FileInfo {
	return FileInfo(name, 0, fs.ModeDir, modTime, "")
}

type staticFileInfo struct {
	modTime     time.Time
	name        string
	contentType string
	size        int64
	mode        fs.FileMode
}

var (
	_ fs.FileInfo = (*staticFileInfo)(nil)
	_ fs.DirEntry = (*staticFileInfo)(nil)
)

func (fi staticFileInfo) ContentType() string         { return fi.contentType }
func (fi staticFileInfo) IsDir() bool                 { return fi.Mode().IsDir() }
func (fi staticFileInfo) Mode() fs.FileMode           { return fi.mode }
func (fi *staticFileInfo) ModTime() time.Time         { return fi.modTime }
func (fi staticFileInfo) Name() string                { return fi.name }
func (fi staticFileInfo) Size() int64                 { return fi.size }
func (fi staticFileInfo) Sys() any                    { return nil }
func (fi *staticFileInfo) Info() (fs.FileInfo, error) { return fi, nil }
func (fi staticFileInfo) Type() fs.FileMode           { return fi.Mode().Type() }

// FileInfoDirEntry adapts a fs.FileInfo into a fs.DirEntry. If it doesn't
// already implement fs.DirEntry, it will be wrapped to always return the
// same fs.FileInfo.
func FileInfoDirEntry(fi fs.FileInfo) fs.DirEntry {
	de, ok := fi.(fs.DirEntry)
	if ok {
		return de
	}

	return &fileinfoDirEntry{fi}
}

// a wrapper to make a fs.FileInfo into an fs.DirEntry
type fileinfoDirEntry struct {
	fs.FileInfo
}

var _ fs.DirEntry = (*fileinfoDirEntry)(nil)

func (fi *fileinfoDirEntry) Info() (fs.FileInfo, error) { return fi, nil }
func (fi *fileinfoDirEntry) Type() fs.FileMode          { return fi.Mode().Type() }
