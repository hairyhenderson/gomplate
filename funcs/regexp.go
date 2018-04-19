package funcs

import (
	"sync"

	"github.com/hairyhenderson/gomplate/conv"
	"github.com/hairyhenderson/gomplate/regexp"
)

var (
	reNS     *ReFuncs
	reNSInit sync.Once
)

// ReNS -
func ReNS() *ReFuncs {
	reNSInit.Do(func() { reNS = &ReFuncs{} })
	return reNS
}

// AddReFuncs -
func AddReFuncs(f map[string]interface{}) {
	f["regexp"] = ReNS
}

// ReFuncs -
type ReFuncs struct{}

// Replace -
func (f *ReFuncs) Replace(re, replacement string, input interface{}) string {
	return regexp.Replace(re, replacement, conv.ToString(input))
}

// Match -
func (f *ReFuncs) Match(re string, input interface{}) bool {
	return regexp.Match(re, conv.ToString(input))
}
