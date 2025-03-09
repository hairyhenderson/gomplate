package gomplate

import (
	"context"
	"maps"
	"text/template"

	"github.com/hairyhenderson/gomplate/v4/internal/config"
	"github.com/hairyhenderson/gomplate/v4/internal/funcs"
)

// CreateFuncs - function mappings are created here
func CreateFuncs(ctx context.Context) template.FuncMap {
	f := template.FuncMap{}
	maps.Copy(f, funcs.CreateDataFuncs(ctx))
	maps.Copy(f, funcs.CreateAWSFuncs(ctx))
	maps.Copy(f, funcs.CreateGCPFuncs(ctx))
	maps.Copy(f, funcs.CreateBase64Funcs(ctx))
	maps.Copy(f, funcs.CreateNetFuncs(ctx))
	maps.Copy(f, funcs.CreateReFuncs(ctx))
	maps.Copy(f, funcs.CreateStringFuncs(ctx))
	maps.Copy(f, funcs.CreateEnvFuncs(ctx))
	maps.Copy(f, funcs.CreateConvFuncs(ctx))
	maps.Copy(f, funcs.CreateTimeFuncs(ctx))
	maps.Copy(f, funcs.CreateMathFuncs(ctx))
	maps.Copy(f, funcs.CreateCryptoFuncs(ctx))
	maps.Copy(f, funcs.CreateFileFuncs(ctx))
	maps.Copy(f, funcs.CreateFilePathFuncs(ctx))
	maps.Copy(f, funcs.CreatePathFuncs(ctx))
	maps.Copy(f, funcs.CreateSockaddrFuncs(ctx))
	maps.Copy(f, funcs.CreateTestFuncs(ctx))
	maps.Copy(f, funcs.CreateCollFuncs(ctx))
	maps.Copy(f, funcs.CreateUUIDFuncs(ctx))
	maps.Copy(f, funcs.CreateRandomFuncs(ctx))
	maps.Copy(f, funcs.CreateSemverFuncs(ctx))
	return f
}

// SetExperimental enables experimental functions and features in the given
// context. This must be done before creating functions. The set of experimental
// features enabled by this is not fixed and will change over time.
func SetExperimental(ctx context.Context) context.Context {
	// This just calls the internal function. This is here to make experimental
	// functions available to external packages.
	return config.SetExperimental(ctx)
}
