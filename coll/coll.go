package coll

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/imdario/mergo"

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

// Merge source maps (srcs) into dst. Precedence is in left-to-right order, with
// the left-most values taking precedence over the right-most.
func Merge(dst map[string]interface{}, srcs ...map[string]interface{}) (map[string]interface{}, error) {
	for _, src := range srcs {
		err := mergo.Merge(&dst, src)
		if err != nil {
			return nil, err
		}
	}
	return dst, nil
}

// Sort a given array or slice. Uses natural sort order if possible. If a
// non-empty key is given and the list elements are maps, this will attempt to
// sort by the values of those entries.
//
// Does not modify the input list.
func Sort(key string, list interface{}) (out []interface{}, err error) {
	if list == nil {
		return nil, nil
	}

	ia, err := interfaceSlice(list)
	if err != nil {
		return nil, err
	}
	// if the types are all the same, we can sort the slice
	if sameTypes(ia) {
		s := make([]interface{}, len(ia))
		// make a copy so the original is unmodified
		copy(s, ia)
		sort.SliceStable(s, func(i, j int) bool {
			return lessThan(key)(s[i], s[j])
		})
		return s, nil
	}
	return ia, nil
}

// lessThan - compare two values of the same type
func lessThan(key string) func(left, right interface{}) bool {
	return func(left, right interface{}) bool {
		val := reflect.Indirect(reflect.ValueOf(left))
		rval := reflect.Indirect(reflect.ValueOf(right))
		switch val.Kind() {
		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
			return val.Int() < rval.Int()
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint, reflect.Uint64:
			return val.Uint() < rval.Uint()
		case reflect.Float32, reflect.Float64:
			return val.Float() < rval.Float()
		case reflect.String:
			return val.String() < rval.String()
		case reflect.MapOf(
			reflect.TypeOf(reflect.String),
			reflect.TypeOf(reflect.Interface),
		).Kind():
			kval := reflect.ValueOf(key)
			if !val.MapIndex(kval).IsValid() {
				return false
			}
			newleft := val.MapIndex(kval).Interface()
			newright := rval.MapIndex(kval).Interface()
			return lessThan("")(newleft, newright)
		case reflect.Struct:
			if !val.FieldByName(key).IsValid() {
				return false
			}
			newleft := val.FieldByName(key).Interface()
			newright := rval.FieldByName(key).Interface()
			return lessThan("")(newleft, newright)
		default:
			// it's not really comparable, so...
			return false
		}
	}
}

func sameTypes(a []interface{}) bool {
	var t reflect.Type
	for _, v := range a {
		if t == nil {
			t = reflect.TypeOf(v)
		}
		if reflect.ValueOf(v).Kind() != t.Kind() {
			return false
		}
	}
	return true
}
