package gomplate

import (
	"context"
	"os"
	"strings"

	"github.com/hairyhenderson/gomplate/v4/internal/datafs"
	"github.com/hairyhenderson/gomplate/v4/internal/parsers"
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
	sr datafs.DataSourceReader,
) (interface{}, error) {
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
