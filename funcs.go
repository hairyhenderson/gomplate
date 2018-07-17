package gomplate

import (
	"text/template"

	"github.com/hairyhenderson/gomplate/data"
	"github.com/hairyhenderson/gomplate/funcs"
)

// initFuncs - The function mappings are defined here!
func initFuncs(d *data.Data) template.FuncMap {
	f := template.FuncMap{}
	funcs.AddDataFuncs(f, d)
	funcs.AWSFuncs(f)
	funcs.AddBase64Funcs(f)
	funcs.AddNetFuncs(f)
	funcs.AddReFuncs(f)
	funcs.AddStringFuncs(f)
	funcs.AddEnvFuncs(f)
	funcs.AddConvFuncs(f)
	funcs.AddTimeFuncs(f)
	funcs.AddMathFuncs(f)
	funcs.AddCryptoFuncs(f)
	funcs.AddFileFuncs(f)
	funcs.AddFilePathFuncs(f)
	funcs.AddPathFuncs(f)
	funcs.AddSockaddrFuncs(f)
	funcs.AddTestFuncs(f)
	return f
}
