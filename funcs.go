package gomplate

import (
	"context"
	"text/template"

	"github.com/hairyhenderson/gomplate/v4/internal/config"
	"github.com/hairyhenderson/gomplate/v4/internal/funcs"
)

// CreateFuncs - function mappings are created here
func CreateFuncs(ctx context.Context) template.FuncMap {
	f := template.FuncMap{}
	addToMap(f, funcs.CreateDataFuncs(ctx))
	addToMap(f, funcs.CreateAWSFuncs(ctx))
	addToMap(f, funcs.CreateGCPFuncs(ctx))
	addToMap(f, funcs.CreateBase64Funcs(ctx))
	addToMap(f, funcs.CreateNetFuncs(ctx))
	addToMap(f, funcs.CreateReFuncs(ctx))
	addToMap(f, funcs.CreateStringFuncs(ctx))
	addToMap(f, funcs.CreateEnvFuncs(ctx))
	addToMap(f, funcs.CreateConvFuncs(ctx))
	addToMap(f, funcs.CreateTimeFuncs(ctx))
	addToMap(f, funcs.CreateMathFuncs(ctx))
	addToMap(f, funcs.CreateCryptoFuncs(ctx))
	addToMap(f, funcs.CreateFileFuncs(ctx))
	addToMap(f, funcs.CreateFilePathFuncs(ctx))
	addToMap(f, funcs.CreatePathFuncs(ctx))
	addToMap(f, funcs.CreateSockaddrFuncs(ctx))
	addToMap(f, funcs.CreateTestFuncs(ctx))
	addToMap(f, funcs.CreateCollFuncs(ctx))
	addToMap(f, funcs.CreateUUIDFuncs(ctx))
	addToMap(f, funcs.CreateRandomFuncs(ctx))
	addToMap(f, funcs.CreateSemverFuncs(ctx))
	return f
}

// addToMap - add src's entries to dst
func addToMap(dst, src map[string]interface{}) {
	for k, v := range src {
		dst[k] = v
	}
}

// SetExperimental enables experimental functions and features in the given
// context. This must be done before creating functions. The set of experimental
// features enabled by this is not fixed and will change over time.
func SetExperimental(ctx context.Context) context.Context {
	// This just calls the internal function. This is here to make experimental
	// functions available to external packages.
	return config.SetExperimental(ctx)
}
