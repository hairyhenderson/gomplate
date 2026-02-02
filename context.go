package gomplate

import (
	"context"

	"github.com/hairyhenderson/gomplate/v5/internal/datafs"
	"github.com/hairyhenderson/gomplate/v5/internal/funcs"
	"github.com/hairyhenderson/gomplate/v5/internal/parsers"
)

// context for templates
type tmplctx map[string]any

// Env - Map environment variables for use in a template
func (c *tmplctx) Env() map[string]string {
	return funcs.EnvFuncs{}.Env()
}

// createTmplContext reads the datasources for the given aliases
func createTmplContext(
	ctx context.Context, aliases []string,
	sr datafs.DataSourceReader,
) (any, error) {
	tctx := &tmplctx{}
	for _, a := range aliases {
		ct, b, err := sr.ReadSource(ctx, a)
		if err != nil {
			return nil, err
		}

		content, err := parsers.ParseData(ct, string(b))
		if err != nil {
			return nil, err
		}

		if a == "." {
			return content, nil
		}

		(*tctx)[a] = content
	}
	return tctx, nil
}
