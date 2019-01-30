package data

import (
	"strings"

	"github.com/hairyhenderson/gomplate/env"
)

func init() {
	addSourceReader("env", readEnv)
}

func readEnv(source *Source, args ...string) (b []byte, err error) {
	n := source.URL.Path
	n = strings.TrimPrefix(n, "/")
	if n != "" {
	} else if n == "" {
		n = source.URL.Opaque
	}

	b = []byte(env.Getenv(n))
	return b, nil
}
