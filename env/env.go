// Package env contains functions that retrieve data from the environment
package env

import (
	"io/fs"
	"os"
	"strings"

	osfs "github.com/hack-pad/hackpadfs/os"
	"github.com/hairyhenderson/gomplate/v4/internal/datafs"
)

// Getenv - retrieves the value of the environment variable named by the key.
// If the variable is unset, but the same variable ending in `_FILE` is set, the
// referenced file will be read into the value.
// Otherwise the provided default (or an emptry string) is returned.
func Getenv(key string, def ...string) string {
	fsys := datafs.WrapWdFS(osfs.NewFS())
	return getenvVFS(fsys, key, def...)
}

// ExpandEnv - like os.ExpandEnv, except supports `_FILE` vars as well
func ExpandEnv(s string) string {
	fsys := datafs.WrapWdFS(osfs.NewFS())
	return expandEnvVFS(fsys, s)
}

// expandEnvVFS -
func expandEnvVFS(fsys fs.FS, s string) string {
	return os.Expand(s, func(s string) string {
		return getenvVFS(fsys, s)
	})
}

// getenvVFS - a convenience function intended for internal use only!
func getenvVFS(fsys fs.FS, key string, def ...string) string {
	val := getenvFile(fsys, key)
	if val == "" && len(def) > 0 {
		return def[0]
	}

	return val
}

func getenvFile(fsys fs.FS, key string) string {
	val := os.Getenv(key)
	if val != "" {
		return val
	}

	p := os.Getenv(key + "_FILE")
	if p != "" {
		val, err := readFile(fsys, p)
		if err != nil {
			return ""
		}
		return strings.TrimSpace(val)
	}

	return ""
}

func readFile(fsys fs.FS, p string) (string, error) {
	b, err := fs.ReadFile(fsys, p)
	if err != nil {
		return "", err
	}

	return string(b), nil
}
