package datafs

import (
	"context"
	"fmt"
	"io/fs"
	"net/url"
	"path"
	"path/filepath"
	"strings"

	"github.com/hairyhenderson/go-fsimpl"
)

type fsProviderCtxKey struct{}

// ContextWithFSProvider returns a context with the given FSProvider. Should
// only be used in tests.
func ContextWithFSProvider(ctx context.Context, fsp fsimpl.FSProvider) context.Context {
	return context.WithValue(ctx, fsProviderCtxKey{}, fsp)
}

// FSProviderFromContext returns the FSProvider from the context, if any
func FSProviderFromContext(ctx context.Context) fsimpl.FSProvider {
	if fsp, ok := ctx.Value(fsProviderCtxKey{}).(fsimpl.FSProvider); ok {
		return fsp
	}

	return nil
}

// ParseSourceURL parses a datasource URL value, which may be '-' (for stdin://),
// or it may be a Windows path (with driver letter and back-slash separators) or
// UNC, or it may be relative. It also might just be a regular absolute URL...
// In all cases it returns a correct URL for the value. It may be a relative URL
// in which case the scheme should be assumed to be 'file'
func ParseSourceURL(value string) (*url.URL, error) {
	if value == "-" {
		value = "stdin://"
	}
	value = filepath.ToSlash(value)
	// handle absolute Windows paths
	volName := ""
	if volName = filepath.VolumeName(value); volName != "" {
		// handle UNCs
		if len(volName) > 2 {
			value = "file:" + value
		} else {
			value = "file:///" + value
		}
	}
	srcURL, err := url.Parse(value)
	if err != nil {
		return nil, err
	}

	if volName != "" && len(srcURL.Path) >= 3 {
		if srcURL.Path[0] == '/' && srcURL.Path[2] == ':' {
			srcURL.Path = srcURL.Path[1:]
		}
	}

	// if it's an absolute path with no scheme, assume it's a file
	if srcURL.Scheme == "" && path.IsAbs(srcURL.Path) {
		srcURL.Scheme = "file"
	}

	return srcURL, nil
}

// FSysForPath returns an [io/fs.FS] for the given path (which may be an URL),
// rooted at /. A [fsimpl.FSProvider] is required to be present in ctx,
// otherwise an error is returned.
func FSysForPath(ctx context.Context, path string) (fs.FS, error) {
	u, err := ParseSourceURL(path)
	if err != nil {
		return nil, err
	}

	fsp := FSProviderFromContext(ctx)
	if fsp == nil {
		return nil, fmt.Errorf("no filesystem provider in context")
	}

	// default to "/" so we have a rooted filesystem for all schemes, but also
	// support volumes on Windows
	origPath := u.Path
	if u.Scheme == "file" || strings.HasSuffix(u.Scheme, "+file") || u.Scheme == "" {
		u.Path, _, err = ResolveLocalPath(origPath)
		if err != nil {
			return nil, fmt.Errorf("resolve local path %q: %w", origPath, err)
		}
		// if this is a drive letter, add a trailing slash
		if u.Path[0] != '/' {
			u.Path += "/"
		}
	}

	fsys, err := fsp.New(u)
	if err != nil {
		return nil, fmt.Errorf("filesystem provider for %q unavailable: %w", path, err)
	}

	return fsys, nil
}

type fsp struct {
	newFunc func(*url.URL) (fs.FS, error)
	schemes []string
}

func (p fsp) Schemes() []string {
	return p.schemes
}

func (p fsp) New(u *url.URL) (fs.FS, error) {
	return p.newFunc(u)
}

// WrappedFSProvider is an FSProvider that returns the given fs.FS
func WrappedFSProvider(fsys fs.FS, schemes ...string) fsimpl.FSProvider {
	return fsp{
		newFunc: func(u *url.URL) (fs.FS, error) { return fsys, nil },
		schemes: schemes,
	}
}
