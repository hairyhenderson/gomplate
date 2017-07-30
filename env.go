package main

import (
	"os"
	"io/ioutil"

	"github.com/blang/vfs"
)

// Env - functions that deal with the environment
type Env struct {
}

// Getenv retrieves the value of the environment variable named by the key.
// It returns the value, or the default (or an emptry string) if the variable is
// not set.
func (e *Env) Getenv(key string, def ...string) string {
	val := os.Getenv(key)
	if val == "" && len(def) > 0 {
		return def[0]
	}

	return val
}

func (e *Env) GetenvFile(fs vfs.Filesystem, key, def string) string {
	val := os.Getenv(key)
	if val != "" {
		return val
	}

	p := os.Getenv(key + "_FILE")
	if p != "" {
		f, err := fs.OpenFile(p, os.O_RDONLY, 0)
		if err != nil {
			return def
		}
		b, err := ioutil.ReadAll(f)
		if err != nil {
			return def
		}
		return string(b)
	}

	return def
}
