package datafs

import (
	"context"
	"io"
	"io/fs"
	"os"

	"github.com/hairyhenderson/gomplate/v4/internal/config"
)

// withContexter is an fs.FS that can be configured with a custom context
// copied from go-fsimpl - see internal/types.go
type withContexter interface {
	WithContext(ctx context.Context) fs.FS
}

type withDataSourceser interface {
	WithDataSources(sources map[string]config.DataSource) fs.FS
}

// WithDataSourcesFS injects a datasource map into the filesystem fs, if the
// filesystem supports it (i.e. has a WithDataSources method). This is used for
// the mergefs filesystem.
func WithDataSourcesFS(sources map[string]config.DataSource, fsys fs.FS) fs.FS {
	if fsys, ok := fsys.(withDataSourceser); ok {
		return fsys.WithDataSources(sources)
	}

	return fsys
}

type stdinCtxKey struct{}

func ContextWithStdin(ctx context.Context, r io.Reader) context.Context {
	return context.WithValue(ctx, stdinCtxKey{}, r)
}

func StdinFromContext(ctx context.Context) io.Reader {
	if r, ok := ctx.Value(stdinCtxKey{}).(io.Reader); ok {
		return r
	}

	return os.Stdin
}
