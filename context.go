package gomplate

import (
	"os"
	"strings"
)

// context for templates
type context struct {
}

// Env - Map environment variables for use in a template
func (c *context) Env() map[string]string {
	env := make(map[string]string)
	for _, i := range os.Environ() {
		sep := strings.Index(i, "=")
		env[i[0:sep]] = i[sep+1:]
	}
	return env
}
