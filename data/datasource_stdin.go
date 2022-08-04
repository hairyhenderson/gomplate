package data

import (
	"context"
	"io"
	"os"

	"github.com/pkg/errors"
)

func readStdin(ctx context.Context, source *Source, args ...string) ([]byte, error) {
	stdin := stdinFromContext(ctx)

	b, err := io.ReadAll(stdin)
	if err != nil {
		return nil, errors.Wrapf(err, "Can't read %s", stdin)
	}
	return b, nil
}

type stdinCtxKey struct{}

func ContextWithStdin(ctx context.Context, r io.Reader) context.Context {
	return context.WithValue(ctx, stdinCtxKey{}, r)
}

func stdinFromContext(ctx context.Context) io.Reader {
	if r, ok := ctx.Value(stdinCtxKey{}).(io.Reader); ok {
		return r
	}

	return os.Stdin
}
