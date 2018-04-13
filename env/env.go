package env

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/blang/vfs"
)

// Getenv - retrieves the value of the environment variable named by the key.
// If the variable is unset, but the same variable ending in `_FILE` is set, the
// referenced file will be read into the value.
// Otherwise the provided default (or an emptry string) is returned.
func Getenv(key string, def ...string) string {
	return GetenvVFS(vfs.OS(), key, def...)
}

// ExpandEnv - like os.ExpandEnv, except supports `_FILE` vars as well
func ExpandEnv(s string) string {
	return expandEnvVFS(vfs.OS(), s)
}

// expandEnvVFS -
func expandEnvVFS(fs vfs.Filesystem, s string) string {
	return os.Expand(s, func(s string) string {
		return GetenvVFS(fs, s)
	})
}

// GetenvVFS - a convenience function intended for internal use only!
func GetenvVFS(fs vfs.Filesystem, key string, def ...string) string {
	val := getenvFile(fs, key)
	if val == "" && len(def) > 0 {
		return def[0]
	}

	return val
}

func getenvFile(fs vfs.Filesystem, key string) string {
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

func readFile(fs vfs.Filesystem, p string) (string, error) {
	f, err := fs.OpenFile(p, os.O_RDONLY, 0)
	if err != nil {
		return "", err
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
