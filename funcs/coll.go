package funcs

import (
	"sync"

	"github.com/hairyhenderson/gomplate/v3/conv"

	"github.com/hairyhenderson/gomplate/v3/coll"
	"github.com/pkg/errors"
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
	f["merge"] = CollNS().Merge
	f["sort"] = CollNS().Sort
	f["jsonpath"] = CollNS().JSONPath
	f["flatten"] = CollNS().Flatten
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

// Merge -
func (f *CollFuncs) Merge(dst map[string]interface{}, src ...map[string]interface{}) (map[string]interface{}, error) {
	return coll.Merge(dst, src...)
}

// Sort -
func (f *CollFuncs) Sort(args ...interface{}) ([]interface{}, error) {
	var (
		key  string
		list interface{}
	)
	if len(args) == 0 || len(args) > 2 {
		return nil, errors.Errorf("wrong number of args: wanted 1 or 2, got %d", len(args))
	}
	if len(args) == 1 {
		list = args[0]
	}
	if len(args) == 2 {
		key = conv.ToString(args[0])
		list = args[1]
	}
	return coll.Sort(key, list)
}

// JSONPath -
func (f *CollFuncs) JSONPath(p string, in interface{}) (interface{}, error) {
	return coll.JSONPath(p, in)
}

// Flatten -
func (f *CollFuncs) Flatten(args ...interface{}) ([]interface{}, error) {
	if len(args) == 0 || len(args) > 2 {
		return nil, errors.Errorf("wrong number of args: wanted 1 or 2, got %d", len(args))
	}
	list := args[0]
	depth := -1
	if len(args) == 2 {
		depth = conv.ToInt(args[0])
		list = args[1]
	}
	return coll.Flatten(list, depth)
}

func pickOmitArgs(args ...interface{}) (map[string]interface{}, []string, error) {
	if len(args) <= 1 {
		return nil, nil, errors.Errorf("wrong number of args: wanted 2 or more, got %d", len(args))
	}

	m, ok := args[len(args)-1].(map[string]interface{})
	if !ok {
		return nil, nil, errors.Errorf("wrong map type: must be map[string]interface{}, got %T", args[len(args)-1])
	}

	keys := make([]string, len(args)-1)
	for i, v := range args[0 : len(args)-1] {
		k, ok := v.(string)
		if !ok {
			return nil, nil, errors.Errorf("wrong key type: must be string, got %T (%+v)", args[i], args[i])
		}
		keys[i] = k
	}
	return m, keys, nil
}

// Pick -
func (f *CollFuncs) Pick(args ...interface{}) (map[string]interface{}, error) {
	m, keys, err := pickOmitArgs(args...)
	if err != nil {
		return nil, err
	}
	return coll.Pick(m, keys...), nil
}

// Omit -
func (f *CollFuncs) Omit(args ...interface{}) (map[string]interface{}, error) {
	m, keys, err := pickOmitArgs(args...)
	if err != nil {
		return nil, err
	}
	return coll.Omit(m, keys...), nil
}
