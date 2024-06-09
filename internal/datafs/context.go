package datafs

import (
	"context"
	"io"
	"io/fs"
	"os"
)

// withContexter is an fs.FS that can be configured with a custom context
// copied from go-fsimpl - see internal/types.go
type withContexter interface {
	WithContext(ctx context.Context) fs.FS
}

type withDataSourceRegistryer interface {
	WithDataSourceRegistry(registry Registry) fs.FS
}

// WithDataSourceRegistryFS injects a datasource registry into the filesystem fs, if the
// filesystem supports it (i.e. has a WithDataSourceRegistry method). This is used for
// the mergefs filesystem.
func WithDataSourceRegistryFS(registry Registry, fsys fs.FS) fs.FS {
	if fsys, ok := fsys.(withDataSourceRegistryer); ok {
		return fsys.WithDataSourceRegistry(registry)
	}

	return fsys
}

type stdinCtxKey struct{}

// ContextWithStdin injects an [io.Reader] into the context, which can be used
// to override the default stdin.
func ContextWithStdin(ctx context.Context, r io.Reader) context.Context {
	return context.WithValue(ctx, stdinCtxKey{}, r)
}

// StdinFromContext returns the io.Reader that should be used for stdin as
// injected by [ContextWithStdin]. If no reader has been injected, [os.Stdin] is
// returned.
func StdinFromContext(ctx context.Context) io.Reader {
	if r, ok := ctx.Value(stdinCtxKey{}).(io.Reader); ok {
		return r
	}

	return os.Stdin
}
