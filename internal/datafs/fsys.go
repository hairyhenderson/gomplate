package datafs

import (
	"context"
	"fmt"
	"io/fs"
	"net/url"
	"strings"

	"github.com/hairyhenderson/go-fsimpl"
	"github.com/hairyhenderson/go-fsimpl/vaultfs/vaultauth"
	"github.com/hairyhenderson/gomplate/v4/internal/urlhelpers"
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

// FSysForPath returns an [io/fs.FS] for the given path (which may be an URL),
// rooted at /. A [fsimpl.FSProvider] is required to be present in ctx,
// otherwise an error is returned.
func FSysForPath(ctx context.Context, path string) (fs.FS, error) {
	u, err := urlhelpers.ParseSourceURL(path)
	if err != nil {
		return nil, err
	}

	fsp := FSProviderFromContext(ctx)
	if fsp == nil {
		return nil, fmt.Errorf("no filesystem provider in context")
	}

	origPath := u.Path

	switch u.Scheme {
	case "git+file", "git+http", "git+https", "git+ssh", "git":
		// git URLs are special - they have double-slashes that separate a repo from
		// a path in the repo. A missing double-slash means the path is the root.
		u.Path, _, _ = strings.Cut(u.Path, "//")
	}

	switch u.Scheme {
	case "git+http", "git+https", "git+ssh", "git":
		// no-op, these are handled
	case "aws+sm":
		// An aws+sm URL can be opaque, best not disturb it
	case "", "file", "git+file":
		// default to "/" so we have a rooted filesystem for all schemes, but also
		// support volumes on Windows
		root, name, rerr := ResolveLocalPath(nil, u.Path)
		if rerr != nil {
			return nil, fmt.Errorf("resolve local path %q: %w", origPath, rerr)
		}

		// windows absolute paths need a slash between the volume and path
		if root != "" && root[0] != '/' {
			u.Path = root + "/" + name
		} else {
			u.Path = root + name
		}

		// if this is a drive letter, add a trailing slash
		if len(u.Path) == 2 && u.Path[0] != '/' && u.Path[1] == ':' {
			u.Path += "/"
		} else if u.Path[0] != '/' {
			u.Path += "/"
		}

		// if this starts with a drive letter, add a leading slash
		// NOPE - this breaks lots of things
		// if len(u.Path) > 2 && u.Path[0] != '/' && u.Path[1] == ':' {
		// 	u.Path = "/" + u.Path
		// }
	default:
		u.Path = "/"
	}

	fsys, err := fsp.New(u)
	if err != nil {
		return nil, fmt.Errorf("filesystem provider for %q unavailable: %w", path, err)
	}

	// inject vault auth methods if needed
	switch u.Scheme {
	case "vault", "vault+http", "vault+https":
		fileFsys, err := fsp.New(&url.URL{Scheme: "file", Path: "/"})
		if err != nil {
			return nil, fmt.Errorf("filesystem provider for %q unavailable: %w", path, err)
		}
		fsys = vaultauth.WithAuthMethod(compositeVaultAuthMethod(fileFsys), fsys)
	}

	fsys = fsimpl.WithContextFS(ctx, fsys)

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
		newFunc: func(_ *url.URL) (fs.FS, error) { return fsys, nil },
		schemes: schemes,
	}
}
