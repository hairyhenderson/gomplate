package funcs

import (
	"sync"

	"github.com/hairyhenderson/gomplate/env"
)

var (
	ef     *EnvFuncs
	efInit sync.Once
)

// EnvNS - the Env namespace
func EnvNS() *EnvFuncs {
	efInit.Do(func() { ef = &EnvFuncs{} })
	return ef
}

// AddEnvFuncs -
func AddEnvFuncs(f map[string]interface{}) {
	f["env"] = EnvNS

	// global aliases - for backwards compatibility
	f["getenv"] = EnvNS().Getenv
}

// EnvFuncs -
type EnvFuncs struct{}

// Getenv -
func (f *EnvFuncs) Getenv(key string, def ...string) string {
	return env.Getenv(key, def...)
}
