package gencel

import (
	"go/ast"
)

type FuncDecl struct {
	Name        string     `json:"Name"`
	Args        []ast.Expr `json:"Args"`
	ReturnTypes []ast.Expr `json:"ReturnType"`
	Body        string     `json:"Body"`
	RecvType    string     `json:"RecvType"`
}
