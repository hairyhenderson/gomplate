//+build !integration

package integration

import (
	"bytes"
	"context"
	"os"
	"strings"

	gcmd "github.com/hairyhenderson/gomplate/v3/internal/cmd"
)

func (c *command) run() (o, e string, err error) {
	origEnviron := map[string]string{}
	for k, v := range c.env {
		origEnviron[k] = os.Getenv(k)
		os.Setenv(k, v)
	}

	defer func() {
		for k, v := range origEnviron {
			os.Setenv(k, v)
		}
	}()

	if c.dir != "" {
		//nolint:govet
		origWd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		defer func() { os.Chdir(origWd) }()
		err = os.Chdir(c.dir)
		if err != nil {
			panic(err)
		}
	}

	stdin := strings.NewReader(c.stdin)

	ctx := context.Background()
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	err = gcmd.Main(ctx, c.args, stdin, stdout, stderr)
	return stdout.String(), stderr.String(), err
}
