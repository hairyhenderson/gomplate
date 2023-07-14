package gomplate

import (
	"context"
	"text/template"

	"github.com/flanksource/gomplate/v3/funcs" //nolint:staticcheck
)

// CreateFuncs - function mappings are created here
func CreateFuncs(ctx context.Context) template.FuncMap {
	f := template.FuncMap{}
	addToMap(f, funcs.CreateDataFuncs(ctx))
	addToMap(f, funcs.CreateBase64Funcs(ctx))
	addToMap(f, funcs.CreateReFuncs(ctx))
	addToMap(f, funcs.CreateStringFuncs(ctx))
	addToMap(f, funcs.CreateConvFuncs(ctx))
	addToMap(f, funcs.CreateTimeFuncs(ctx))
	addToMap(f, funcs.CreateMathFuncs(ctx))
	addToMap(f, funcs.CreateCryptoFuncs(ctx))
	addToMap(f, funcs.CreateKubernetesFuncs(ctx))
	addToMap(f, funcs.CreateFilePathFuncs(ctx))
	addToMap(f, funcs.CreatePathFuncs(ctx))
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
