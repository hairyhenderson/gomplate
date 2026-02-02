package funcs

import (
	"context"
	"os"
	"strings"

	"github.com/hairyhenderson/gomplate/v5/conv"
	"github.com/hairyhenderson/gomplate/v5/env"
)

// CreateEnvFuncs -
func CreateEnvFuncs(ctx context.Context) map[string]any {
	ns := &EnvFuncs{ctx}

	return map[string]any{
		"env":    func() any { return ns },
		"getenv": ns.Getenv,
	}
}

// EnvFuncs -
type EnvFuncs struct {
	ctx context.Context
}

// Env returns a map of all environment variables
func (EnvFuncs) Env() map[string]string {
	envMap := make(map[string]string)
	for _, e := range os.Environ() {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) != 2 || parts[0] == "" {
			// skip malformed entries or Windows =C: style env vars
			continue
		}
		envMap[parts[0]] = parts[1]
	}
	return envMap
}

// Getenv -
func (EnvFuncs) Getenv(key any, def ...string) string {
	return env.Getenv(conv.ToString(key), def...)
}

// ExpandEnv -
func (EnvFuncs) ExpandEnv(s any) string {
	return env.ExpandEnv(conv.ToString(s))
}

// HasEnv returns true if the environment variable is set, false otherwise
func (EnvFuncs) HasEnv(key any) bool {
	_, ok := os.LookupEnv(conv.ToString(key))
	return ok
}
