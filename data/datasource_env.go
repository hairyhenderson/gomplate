package data

import (
	"context"
	"strings"

	"github.com/hairyhenderson/gomplate/v4/env"
)

//nolint:unparam
func readEnv(_ context.Context, source *Source, _ ...string) (b []byte, err error) {
	n := source.URL.Path
	n = strings.TrimPrefix(n, "/")
	if n == "" {
		n = source.URL.Opaque
	}

	b = []byte(env.Getenv(n))
	return b, nil
}
