package gencel

import (
	"fmt"
	"go/ast"
)

type Ident struct {
	Type       string
	GoType     string
	IsEllipsis bool
}

// getCelArgs converts native go types to cel-go types
func getCelArgs(args []ast.Expr) []Ident {
	var celArgs = make([]Ident, len(args))
	for i, a := range args {
		celArgs[i] = astToIdent(a)
	}

	return celArgs
}

func goTypeToIdent(name string) Ident {
	switch name {
	case "string":
		return Ident{Type: "cel.StringType", GoType: name}
	case "bool":
		return Ident{Type: "cel.BoolType", GoType: name}
	case "Duration":
		return Ident{Type: "cel.DurationType", GoType: "time.Duration"}
	case "Time":
		return Ident{Type: "cel.TimestampType", GoType: "time.Time"}
	case "int", "int8", "int16", "int32", "int64":
		return Ident{Type: "cel.IntType", GoType: name}
	case "uint", "uint8", "uint16", "uint32", "uint64":
		return Ident{Type: "cel.UintType", GoType: name}
	case "float32", "float64":
		return Ident{Type: "cel.DoubleType", GoType: name}
	case "[]byte":
		return Ident{Type: "cel.BytesType", GoType: name}
	case "map":
		return Ident{Type: "cel.MapType", GoType: name}
	case "array", "slice":
		return Ident{Type: "cel.ListType", GoType: name}
	default:
		return Ident{Type: "cel.DynType", GoType: name}
	}
}

func astToIdent(a ast.Expr) Ident {
	switch v := a.(type) {
	case *ast.InterfaceType:
		return Ident{Type: "cel.DynType", GoType: "interface{}"}
	case *ast.Ident:
		return goTypeToIdent(v.Name)
	case *ast.Ellipsis:
		return Ident{Type: "cel.DynType", IsEllipsis: true, GoType: astToIdent(v.Elt).GoType}
	case *ast.SelectorExpr:
		return goTypeToIdent(fmt.Sprintf("%s.%s", astToIdent(v.X).GoType, v.Sel.Name))
	case *ast.ArrayType:
		return Ident{Type: "cel.DynType", GoType: fmt.Sprintf("[]%s", astToIdent(v.Elt).GoType)}
	case *ast.MapType:
		return Ident{Type: "cel.DynType", GoType: fmt.Sprintf("map[%s]%s", astToIdent(v.Key).GoType, astToIdent(v.Value).GoType)}
	default:
		return Ident{Type: "cel.Unknown", GoType: "Unknown"}
	}
}
