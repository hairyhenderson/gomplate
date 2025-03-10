package funcs

import (
	"context"
	"path"

	"github.com/hairyhenderson/gomplate/v4/conv"
)

// CreatePathFuncs -
func CreatePathFuncs(ctx context.Context) map[string]any {
	ns := &PathFuncs{ctx}
	return map[string]any{
		"path": func() any { return ns },
	}
}

// PathFuncs -
type PathFuncs struct {
	ctx context.Context
}

// Base -
func (PathFuncs) Base(in any) string {
	return path.Base(conv.ToString(in))
}

// Clean -
func (PathFuncs) Clean(in any) string {
	return path.Clean(conv.ToString(in))
}

// Dir -
func (PathFuncs) Dir(in any) string {
	return path.Dir(conv.ToString(in))
}

// Ext -
func (PathFuncs) Ext(in any) string {
	return path.Ext(conv.ToString(in))
}

// IsAbs -
func (PathFuncs) IsAbs(in any) bool {
	return path.IsAbs(conv.ToString(in))
}

// Join -
func (PathFuncs) Join(elem ...any) string {
	s := conv.ToStrings(elem...)
	return path.Join(s...)
}

// Match -
func (PathFuncs) Match(pattern, name any) (matched bool, err error) {
	return path.Match(conv.ToString(pattern), conv.ToString(name))
}

// Split -
func (PathFuncs) Split(in any) []string {
	dir, file := path.Split(conv.ToString(in))
	return []string{dir, file}
}
