package main

import "os"

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
