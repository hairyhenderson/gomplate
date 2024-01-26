package funcs

import (
	"context"
	"path"

	"github.com/hairyhenderson/gomplate/v4/conv"
)

// CreatePathFuncs -
func CreatePathFuncs(ctx context.Context) map[string]interface{} {
	ns := &PathFuncs{ctx}
	return map[string]interface{}{
		"path": func() interface{} { return ns },
	}
}

// PathFuncs -
type PathFuncs struct {
	ctx context.Context
}

// Base -
func (PathFuncs) Base(in interface{}) string {
	return path.Base(conv.ToString(in))
}

// Clean -
func (PathFuncs) Clean(in interface{}) string {
	return path.Clean(conv.ToString(in))
}

// Dir -
func (PathFuncs) Dir(in interface{}) string {
	return path.Dir(conv.ToString(in))
}

// Ext -
func (PathFuncs) Ext(in interface{}) string {
	return path.Ext(conv.ToString(in))
}

// IsAbs -
func (PathFuncs) IsAbs(in interface{}) bool {
	return path.IsAbs(conv.ToString(in))
}

// Join -
func (PathFuncs) Join(elem ...interface{}) string {
	s := conv.ToStrings(elem...)
	return path.Join(s...)
}

// Match -
func (PathFuncs) Match(pattern, name interface{}) (matched bool, err error) {
	return path.Match(conv.ToString(pattern), conv.ToString(name))
}

// Split -
func (PathFuncs) Split(in interface{}) []string {
	dir, file := path.Split(conv.ToString(in))
	return []string{dir, file}
}
