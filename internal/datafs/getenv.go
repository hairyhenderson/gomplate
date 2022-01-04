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
