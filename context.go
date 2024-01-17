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
func createTmplContext(
	ctx context.Context, aliases []string,
	//nolint:staticcheck
	d *data.Data,
) (interface{}, error) {
	// we need to inject the current context into the Data value, because
	// the Datasource method may need it
	// TODO: remove this before v4
	if d != nil {
		d.Ctx = ctx
	}

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
