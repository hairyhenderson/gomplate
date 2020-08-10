package gomplate

import (
	"context"
	"text/template"

	"github.com/hairyhenderson/gomplate/v3/data"
	"github.com/hairyhenderson/gomplate/v3/funcs"
	"github.com/hairyhenderson/gomplate/v3/internal/config"
)

// Funcs -
// Deprecated: use CreateFuncs instead
func Funcs(d *data.Data) template.FuncMap {
	ctx := context.Background()
	cfg := config.FromContext(ctx)
	return CreateFuncs(config.ContextWithConfig(ctx, cfg), d)
}

// CreateFuncs - function mappings are created here
func CreateFuncs(ctx context.Context, d *data.Data) template.FuncMap {
	f := template.FuncMap{}
	addToMap(f, funcs.CreateDataFuncs(ctx, d))
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
	return f
}

// addToMap - add src's entries to dst
func addToMap(dst, src map[string]interface{}) {
	for k, v := range src {
		dst[k] = v
	}
}
