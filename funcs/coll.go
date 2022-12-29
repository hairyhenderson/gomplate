package funcs

import (
	"context"

	"github.com/flanksource/gomplate/v3/conv"

	"github.com/flanksource/gomplate/v3/coll"
	"github.com/pkg/errors"
)

// CollNS -
// Deprecated: don't use
func CollNS() *CollFuncs {
	return &CollFuncs{}
}

// AddCollFuncs -
// Deprecated: use CreateCollFuncs instead
func AddCollFuncs(f map[string]interface{}) {
	for k, v := range CreateCollFuncs(context.Background()) {
		f[k] = v
	}
}

// CreateCollFuncs -
func CreateCollFuncs(ctx context.Context) map[string]interface{} {
	f := map[string]interface{}{}

	ns := &CollFuncs{ctx}
	f["coll"] = func() interface{} { return ns }

	f["has"] = ns.Has
	f["slice"] = ns.Slice
	f["dict"] = ns.Dict
	f["keys"] = ns.Keys
	f["values"] = ns.Values
	f["append"] = ns.Append
	f["prepend"] = ns.Prepend
	f["uniq"] = ns.Uniq
	f["reverse"] = ns.Reverse
	f["merge"] = ns.Merge
	f["sort"] = ns.Sort
	f["jsonpath"] = ns.JSONPath
	f["jq"] = ns.JQ
	f["flatten"] = ns.Flatten
	return f
}

// CollFuncs -
type CollFuncs struct {
	ctx context.Context
}

// Slice -
func (CollFuncs) Slice(args ...interface{}) []interface{} {
	return coll.Slice(args...)
}

// Has -
func (CollFuncs) Has(in interface{}, key string) bool {
	return coll.Has(in, key)
}

// Dict -
func (CollFuncs) Dict(in ...interface{}) (map[string]interface{}, error) {
	return coll.Dict(in...)
}

// Keys -
func (CollFuncs) Keys(in ...map[string]interface{}) ([]string, error) {
	return coll.Keys(in...)
}

// Values -
func (CollFuncs) Values(in ...map[string]interface{}) ([]interface{}, error) {
	return coll.Values(in...)
}

// Append -
func (CollFuncs) Append(v interface{}, list interface{}) ([]interface{}, error) {
	return coll.Append(v, list)
}

// Prepend -
func (CollFuncs) Prepend(v interface{}, list interface{}) ([]interface{}, error) {
	return coll.Prepend(v, list)
}

// Uniq -
func (CollFuncs) Uniq(in interface{}) ([]interface{}, error) {
	return coll.Uniq(in)
}

// Reverse -
func (CollFuncs) Reverse(in interface{}) ([]interface{}, error) {
	return coll.Reverse(in)
}

// Merge -
func (CollFuncs) Merge(dst map[string]interface{}, src ...map[string]interface{}) (map[string]interface{}, error) {
	return coll.Merge(dst, src...)
}

// Sort -
func (CollFuncs) Sort(args ...interface{}) ([]interface{}, error) {
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
func (CollFuncs) JSONPath(p string, in interface{}) (interface{}, error) {
	return coll.JSONPath(p, in)
}

// JQ -
func (f *CollFuncs) JQ(jqExpr string, in interface{}) (interface{}, error) {
	return coll.JQ(f.ctx, jqExpr, in)
}

// Flatten -
func (CollFuncs) Flatten(args ...interface{}) ([]interface{}, error) {
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
func (CollFuncs) Pick(args ...interface{}) (map[string]interface{}, error) {
	m, keys, err := pickOmitArgs(args...)
	if err != nil {
		return nil, err
	}
	return coll.Pick(m, keys...), nil
}

// Omit -
func (CollFuncs) Omit(args ...interface{}) (map[string]interface{}, error) {
	m, keys, err := pickOmitArgs(args...)
	if err != nil {
		return nil, err
	}
	return coll.Omit(m, keys...), nil
}
