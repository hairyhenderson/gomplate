package funcs

import (
	"context"

	"github.com/hairyhenderson/gomplate/v4/conv"
	"github.com/hairyhenderson/gomplate/v4/env"
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

// Getenv -
func (EnvFuncs) Getenv(key any, def ...string) string {
	return env.Getenv(conv.ToString(key), def...)
}

// ExpandEnv -
func (EnvFuncs) ExpandEnv(s any) string {
	return env.ExpandEnv(conv.ToString(s))
}
