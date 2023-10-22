package gencel

import (
	"fmt"
	"strings"
)

func getArgs(args []Ident) string {
	output := make([]string, 0, len(args))
	for i := range args {
		var a string
		if args[i].IsEllipsis {
			a = "list..."
			output = append(output, a)
			continue
		}

		switch args[i].GoType {
		case "interface{}":
			a = fmt.Sprintf("args[%d]", i)
		default:
			a = fmt.Sprintf("args[%d].Value().(%s)", i, args[i].GoType)
		}

		output = append(output, a)
	}

	return strings.Join(output, ", ")
}

var tplFuncs = map[string]any{
	"getReturnIdentifiers": func(args []Ident) string {
		var output []string
		for i := range args {
			output = append(output, fmt.Sprintf("a%d", i))
		}

		return strings.Join(output, ", ")
	},
	"fnSuffix": func(args []Ident) string {
		var output []string
		for _, a := range args {
			output = append(output, a.GoType)
		}

		return strings.Join(output, "_")
	},
	"getArgs": getArgs,
	"getReturnTypes": func(args []Ident) string {
		switch len(args) {
		case 0:
			return "nil"
		case 1:
			return goTypeToIdent(args[0].GoType).Type
		default:
			return "cel.DynType"
		}
	},
}

type VariadicArg struct {
	Pos int

	Type string
}

func getVariadicArg(items []Ident) *VariadicArg {
	for i, a := range items {
		if a.IsEllipsis {
			return &VariadicArg{Pos: i, Type: a.GoType}
		}
	}

	return nil
}

type funcDefTemplateView struct {
	// IdentName is the name of the exported cel func
	// in this codebase.
	IdentName string

	// FnName is the name of the cel func inside the
	// cel environment.
	FnName string

	// FnNameWithNamespace is the name of the cel func inside the
	// cel environment that's namespaced by the filename.
	// Example: A function IsMax() inside the file math.go would have
	// FnNameWithNamespace as math.IsMax.
	//
	// NOTE: There are some exceptions so not all function names are namespaced.
	FnNameWithNamespace string

	// Args is the list of arguments of the go func
	// that this cel func is encapsulating.
	Args []Ident

	// ReturnTypes is the list of all returns of the go func
	// that this cel func is encapsulating.
	ReturnTypes []Ident

	// RecvType is the parent type of the member func
	// that this cel func is encapsulating.
	RecvType string

	// VariadicArg indicates whether this func has any ellipsis argument.
	VariadicArg *VariadicArg
}

const funcBodyTemplate = `
{{define "body"}}
		var x {{.RecvType}}
		{{if .VariadicArg}}list := sliceToNative[{{.VariadicArg.Type}}](args[{{.VariadicArg.Pos}}].(ref.Val)){{end}}
		{{if gt (len .ReturnTypes) 1}}
			{{getReturnIdentifiers .ReturnTypes}} := x.{{.FnName}}({{getArgs .Args}})
			return types.DefaultTypeAdapter.NativeToValue([]any{
				{{getReturnIdentifiers .ReturnTypes}},
			})
		{{else}}
			return types.DefaultTypeAdapter.NativeToValue(x.{{.FnName}}({{getArgs .Args}}))
		{{end}}
{{end}}
`

const funcDefTemplate = `
var {{.IdentName}} = cel.Function("{{.FnNameWithNamespace}}",
	cel.Overload("{{.FnNameWithNamespace}}_{{fnSuffix .Args}}",
	{{if .Args}}
	[]*cel.Type{
		{{ range $elem := .Args }} {{.Type}},	{{end}}
	}{{else}}nil{{end}},
	{{getReturnTypes .ReturnTypes}},
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {
			{{ block "body" . }}{{end}}
		}),
	),
)
`

type exportFuncsTemplateView struct {
	FnNames []string
}

const exportAllTemplate = `
func sliceToNative[K any](arg ref.Val) []K {
	list, ok := arg.Value().([]ref.Val)
	if !ok {
		log.Printf("Not a list %T\n", arg.Value())
		return nil
	}

	var out = make([]K, len(list))
	for i, val := range list {
		out[i] = val.Value().(K)
	}

	return out
}

var CelEnvOption = []cel.EnvOption{
	{{range $fnName := .FnNames}}{{$fnName}},
	{{end}}
}
`
