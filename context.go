package gomplate

import (
	"context"
	"os"
	"strings"

	"github.com/hairyhenderson/gomplate/v4/data"
)

// context for templates
type tmplctx map[string]interface{}

// Env - Map environment variables for use in a template
func (c *tmplctx) Env() map[string]string {
	env := make(map[string]string)
	for _, i := range os.Environ() {
		sep := strings.Index(i, "=")
		env[i[0:sep]] = i[sep+1:]
	}
	return env
}

// createTmplContext reads the datasources for the given aliases
//
//nolint:staticcheck
func createTmplContext(_ context.Context, aliases []string, d *data.Data) (interface{}, error) {
	var err error
	tctx := &tmplctx{}
	for _, a := range aliases {
		if a == "." {
			return d.Datasource(a)
		}
		(*tctx)[a], err = d.Datasource(a)
		if err != nil {
			return nil, err
		}
	}
	return tctx, nil
}
