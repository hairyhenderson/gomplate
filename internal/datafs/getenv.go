package datafs

import (
	"io/fs"
	"os"
	"strings"
)

// ExpandEnvFsys - a convenience function intended for internal use only!
func ExpandEnvFsys(fsys fs.FS, s string) string {
	return os.Expand(s, func(s string) string {
		return GetenvFsys(fsys, s)
	})
}

// GetenvFsys - a convenience function intended for internal use only!
func GetenvFsys(fsys fs.FS, key string, def ...string) string {
	val, _ := getenvFile(fsys, key)
	if val == "" && len(def) > 0 {
		return def[0]
	}

	return val
}

// LookupEnvFsys - a convenience function intended for internal use only!
func LookupEnvFsys(fsys fs.FS, key string) (string, bool) {
	return getenvFile(fsys, key)
}

func getenvFile(fsys fs.FS, key string) (string, bool) {
	val, found := os.LookupEnv(key)
	if val != "" {
		return val, true
	}

	p := os.Getenv(key + "_FILE")
	if p != "" {
		val, err := readFile(fsys, p)
		if err != nil {
			return "", false
		}

		return strings.TrimSpace(val), true
	}

	return "", found
}

func readFile(fsys fs.FS, p string) (string, error) {
	b, err := fs.ReadFile(fsys, p)
	if err != nil {
		return "", err
	}

	return string(b), nil
}
