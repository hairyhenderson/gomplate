// Package env contains functions that retrieve data from the environment
package env

import (
	osfs "github.com/hack-pad/hackpadfs/os"
	"github.com/hairyhenderson/gomplate/v4/internal/datafs"
)

// Getenv - retrieves the value of the environment variable named by the key.
// If the variable is unset, but the same variable ending in `_FILE` is set, the
// referenced file will be read into the value.
// Otherwise the provided default (or an empty string) is returned.
func Getenv(key string, def ...string) string {
	fsys := datafs.WrapWdFS(osfs.NewFS())
	return datafs.GetenvFsys(fsys, key, def...)
}

// ExpandEnv - like os.ExpandEnv, except supports `_FILE` vars as well
func ExpandEnv(s string) string {
	fsys := datafs.WrapWdFS(osfs.NewFS())
	return datafs.ExpandEnvFsys(fsys, s)
}

// LookupEnv - retrieves the value of the environment variable named by the key.
// If the variable is unset, but the same variable ending in `_FILE` is set, the
// referenced file will be read into the value. If the key is not set, the
// second return value will be false.
// Otherwise the provided default (or an empty string) is returned.
func LookupEnv(key string) (string, bool) {
	fsys := datafs.WrapWdFS(osfs.NewFS())
	return datafs.LookupEnvFsys(fsys, key)
}
