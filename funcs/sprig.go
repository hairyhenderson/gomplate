package funcs

import (
	"context"
	"github.com/Masterminds/sprig"
	"text/template"
)

func AddSprigFuncs(ctx context.Context, funcs template.FuncMap) {
	sprigFuncs := sprig.TxtFuncMap()
	for funcName := range sprigFuncs {
		if _, ok := funcs[funcName]; !ok {
			funcs[funcName] = sprigFuncs[funcName]
		}
	}
}
