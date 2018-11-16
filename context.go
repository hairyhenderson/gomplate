package gomplate

import (
	"os"
	"strings"

	"github.com/hairyhenderson/gomplate/data"
)

// context for templates
type context map[string]interface{}

// Env - Map environment variables for use in a template
func (c *context) Env() map[string]string {
	env := make(map[string]string)
	for _, i := range os.Environ() {
		sep := strings.Index(i, "=")
		env[i[0:sep]] = i[sep+1:]
	}
	return env
}

func createContext(contexts []string, d *data.Data) (interface{}, error) {
	var err error
	ctx := &context{}
	for _, context := range contexts {
		a := parseAlias(context)
		if a == "." {
			return d.Datasource(a)
		}
		(*ctx)[a], err = d.Datasource(a)
		if err != nil {
			return nil, err
		}
	}
	return ctx, nil
}

func parseAlias(arg string) string {
	parts := strings.SplitN(arg, "=", 2)
	switch len(parts) {
	case 1:
		return strings.SplitN(parts[0], ".", 2)[0]
	default:
		return parts[0]
	}
}
