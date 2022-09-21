package funcs

import (
	"context"

	"github.com/flanksource/gomplate/v3/conv"
	"github.com/flanksource/gomplate/v3/env"
)

// EnvNS - the Env namespace
// Deprecated: don't use
func EnvNS() *EnvFuncs {
	return &EnvFuncs{}
}

// AddEnvFuncs -
// Deprecated: use CreateEnvFuncs instead
func AddEnvFuncs(f map[string]interface{}) {
	for k, v := range CreateEnvFuncs(context.Background()) {
		f[k] = v
	}
}

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
