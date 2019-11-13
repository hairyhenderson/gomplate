package funcs

import (
	"path"
	"sync"

	"github.com/hairyhenderson/gomplate/v3/conv"
)

var (
	pf     *PathFuncs
	pfInit sync.Once
)

// PathNS - the Path namespace
func PathNS() *PathFuncs {
	pfInit.Do(func() { pf = &PathFuncs{} })
	return pf
}

// AddPathFuncs -
func AddPathFuncs(f map[string]interface{}) {
	f["path"] = PathNS
}

// PathFuncs -
type PathFuncs struct {
}

// Base -
func (f *PathFuncs) Base(in interface{}) string {
	return path.Base(conv.ToString(in))
}

// Clean -
func (f *PathFuncs) Clean(in interface{}) string {
	return path.Clean(conv.ToString(in))
}

// Dir -
func (f *PathFuncs) Dir(in interface{}) string {
	return path.Dir(conv.ToString(in))
}

// Ext -
func (f *PathFuncs) Ext(in interface{}) string {
	return path.Ext(conv.ToString(in))
}

// IsAbs -
func (f *PathFuncs) IsAbs(in interface{}) bool {
	return path.IsAbs(conv.ToString(in))
}

// Join -
func (f *PathFuncs) Join(elem ...interface{}) string {
	s := conv.ToStrings(elem...)
	return path.Join(s...)
}

// Match -
func (f *PathFuncs) Match(pattern, name interface{}) (matched bool, err error) {
	return path.Match(conv.ToString(pattern), conv.ToString(name))
}

// Split -
func (f *PathFuncs) Split(in interface{}) []string {
	dir, file := path.Split(conv.ToString(in))
	return []string{dir, file}
}
