package funcs

import (
	"context"

	"github.com/hairyhenderson/gomplate/v4/conv"
	"github.com/hairyhenderson/gomplate/v4/env"
)

// CreateEnvFuncs -
func CreateEnvFuncs(ctx context.Context) map[string]interface{} {
	ns := &EnvFuncs{ctx}

	return map[string]interface{}{
		"env":    func() interface{} { return ns },
		"getenv": ns.Getenv,
	}
}

// EnvFuncs -
type EnvFuncs struct {
	ctx context.Context
}

// Getenv -
func (EnvFuncs) Getenv(key interface{}, def ...string) string {
	return env.Getenv(conv.ToString(key), def...)
}

// ExpandEnv -
func (EnvFuncs) ExpandEnv(s interface{}) string {
	return env.ExpandEnv(conv.ToString(s))
}
