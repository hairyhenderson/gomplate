// Package env contains functions that retrieve data from the environment
package env

import (
	"io"
	"os"
	"strings"

	"github.com/spf13/afero"
)

// Getenv - retrieves the value of the environment variable named by the key.
// If the variable is unset, but the same variable ending in `_FILE` is set, the
// referenced file will be read into the value.
// Otherwise the provided default (or an emptry string) is returned.
func Getenv(key string, def ...string) string {
	return getenvVFS(afero.NewOsFs(), key, def...)
}

// ExpandEnv - like os.ExpandEnv, except supports `_FILE` vars as well
func ExpandEnv(s string) string {
	return expandEnvVFS(afero.NewOsFs(), s)
}

// expandEnvVFS -
func expandEnvVFS(fs afero.Fs, s string) string {
	return os.Expand(s, func(s string) string {
		return getenvVFS(fs, s)
	})
}

// getenvVFS - a convenience function intended for internal use only!
func getenvVFS(fs afero.Fs, key string, def ...string) string {
	val := getenvFile(fs, key)
	if val == "" && len(def) > 0 {
		return def[0]
	}

	return val
}

func getenvFile(fs afero.Fs, key string) string {
	val := os.Getenv(key)
	if val != "" {
		return val
	}

	p := os.Getenv(key + "_FILE")
	if p != "" {
		val, err := readFile(fs, p)
		if err != nil {
			return ""
		}
		return strings.TrimSpace(val)
	}

	return ""
}

func readFile(fs afero.Fs, p string) (string, error) {
	f, err := fs.OpenFile(p, os.O_RDONLY, 0)
	if err != nil {
		return "", err
	}
	b, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
