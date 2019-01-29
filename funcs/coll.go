package funcs

import (
	"sync"

	"github.com/hairyhenderson/gomplate/coll"
)

var (
	collNS     *CollFuncs
	collNSInit sync.Once
)

// CollNS -
func CollNS() *CollFuncs {
	collNSInit.Do(func() { collNS = &CollFuncs{} })
	return collNS
}

// AddCollFuncs -
func AddCollFuncs(f map[string]interface{}) {
	f["coll"] = CollNS

	f["has"] = CollNS().Has
	f["slice"] = CollNS().Slice
	f["dict"] = CollNS().Dict
	f["keys"] = CollNS().Keys
	f["values"] = CollNS().Values
	f["append"] = CollNS().Append
	f["prepend"] = CollNS().Prepend
	f["uniq"] = CollNS().Uniq
	f["reverse"] = CollNS().Reverse
}

// CollFuncs -
type CollFuncs struct{}

// Slice -
func (f *CollFuncs) Slice(args ...interface{}) []interface{} {
	return coll.Slice(args...)
}

// Has -
func (f *CollFuncs) Has(in interface{}, key string) bool {
	return coll.Has(in, key)
}

// Dict -
func (f *CollFuncs) Dict(in ...interface{}) (map[string]interface{}, error) {
	return coll.Dict(in...)
}

// Keys -
func (f *CollFuncs) Keys(in ...map[string]interface{}) ([]string, error) {
	return coll.Keys(in...)
}

// Values -
func (f *CollFuncs) Values(in ...map[string]interface{}) ([]interface{}, error) {
	return coll.Values(in...)
}

// Append -
func (f *CollFuncs) Append(v interface{}, list interface{}) ([]interface{}, error) {
	return coll.Append(v, list)
}

// Prepend -
func (f *CollFuncs) Prepend(v interface{}, list interface{}) ([]interface{}, error) {
	return coll.Prepend(v, list)
}

// Uniq -
func (f *CollFuncs) Uniq(in interface{}) ([]interface{}, error) {
	return coll.Uniq(in)
}

// Reverse -
func (f *CollFuncs) Reverse(in interface{}) ([]interface{}, error) {
	return coll.Reverse(in)
}
