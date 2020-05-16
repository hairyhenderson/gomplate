package datasource

import (
	"context"
	"net/url"
	"strings"

	"github.com/hairyhenderson/gomplate/v3/env"
)

// Env -
type Env struct {
}

var _ Reader = (*Env)(nil)

func (b *Env) Read(ctx context.Context, url *url.URL, args ...string) (data Data, err error) {
	n := url.Path
	n = strings.TrimPrefix(n, "/")
	if n != "" {
	} else if n == "" {
		n = url.Opaque
	}

	data.Bytes = []byte(env.Getenv(n))
	return data, nil
}
