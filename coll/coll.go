package coll

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/hairyhenderson/gomplate/conv"
	"github.com/pkg/errors"
)

// Slice creates a slice from a bunch of arguments
func Slice(args ...interface{}) []interface{} {
	return args
}

func interfaceSlice(slice interface{}) ([]interface{}, error) {
	s := reflect.ValueOf(slice)
	kind := s.Kind()
	switch kind {
	case reflect.Slice, reflect.Array:
		ret := make([]interface{}, s.Len())
		for i := 0; i < s.Len(); i++ {
			ret[i] = s.Index(i).Interface()
		}
		return ret, nil
	default:
		return nil, errors.Errorf("expected an array or slice, but got a %T", s)
	}
}

// Has determines whether or not a given object has a property with the given key
func Has(in interface{}, key interface{}) bool {
	av := reflect.ValueOf(in)

	switch av.Kind() {
	case reflect.Map:
		kv := reflect.ValueOf(key)
		return av.MapIndex(kv).IsValid()
	case reflect.Slice, reflect.Array:
		l := av.Len()
		for i := 0; i < l; i++ {
			v := av.Index(i).Interface()
			if reflect.DeepEqual(v, key) {
				return true
			}
		}
	}

	return false
}

// Dict is a convenience function that creates a map with string keys.
// Provide arguments as key/value pairs. If an odd number of arguments
// is provided, the last is used as the key, and an empty string is
// set as the value.
// All keys are converted to strings, regardless of input type.
func Dict(v ...interface{}) (map[string]interface{}, error) {
	dict := map[string]interface{}{}
	lenv := len(v)
	for i := 0; i < lenv; i += 2 {
		key := conv.ToString(v[i])
		if i+1 >= lenv {
			dict[key] = ""
			continue
		}
		dict[key] = v[i+1]
	}
	return dict, nil
}

// Keys returns the list of keys in one or more maps. The returned list of keys
// is ordered by map, each in sorted key order.
func Keys(in ...map[string]interface{}) ([]string, error) {
	if len(in) == 0 {
		return nil, fmt.Errorf("need at least one argument")
	}
	keys := []string{}
	for _, m := range in {
		k, _ := splitMap(m)
		keys = append(keys, k...)
	}
	return keys, nil
}

func splitMap(m map[string]interface{}) ([]string, []interface{}) {
	keys := make([]string, len(m))
	values := make([]interface{}, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	for i, k := range keys {
		values[i] = m[k]
	}
	return keys, values
}

// Values returns the list of values in one or more maps. The returned list of values
// is ordered by map, each in sorted key order. If the Keys function is called with
// the same arguments, the key/value mappings will be maintained.
func Values(in ...map[string]interface{}) ([]interface{}, error) {
	if len(in) == 0 {
		return nil, fmt.Errorf("need at least one argument")
	}
	values := []interface{}{}
	for _, m := range in {
		_, v := splitMap(m)
		values = append(values, v...)
	}
	return values, nil
}

// Append v to the end of list. No matter what type of input slice or array list is, a new []interface{} is always returned.
func Append(v interface{}, list interface{}) ([]interface{}, error) {
	l, err := interfaceSlice(list)
	if err != nil {
		return nil, err
	}

	return append(l, v), nil
}

// Prepend v to the beginning of list. No matter what type of input slice or array list is, a new []interface{} is always returned.
func Prepend(v interface{}, list interface{}) ([]interface{}, error) {
	l, err := interfaceSlice(list)
	if err != nil {
		return nil, err
	}

	return append([]interface{}{v}, l...), nil
}

// Uniq finds the unique values within list. No matter what type of input slice or array list is, a new []interface{} is always returned.
func Uniq(list interface{}) ([]interface{}, error) {
	l, err := interfaceSlice(list)
	if err != nil {
		return nil, err
	}

	out := []interface{}{}
	for _, v := range l {
		if !Has(out, v) {
			out = append(out, v)
		}
	}
	return out, nil
}

// Reverse the list. No matter what type of input slice or array list is, a new []interface{} is always returned.
func Reverse(list interface{}) ([]interface{}, error) {
	l, err := interfaceSlice(list)
	if err != nil {
		return nil, err
	}

	// nifty trick from https://github.com/golang/go/wiki/SliceTricks#reversing
	for left, right := 0, len(l)-1; left < right; left, right = left+1, right-1 {
		l[left], l[right] = l[right], l[left]
	}
	return l, nil
}
