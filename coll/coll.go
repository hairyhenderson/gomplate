// Package coll contains functions to help manipulate and query collections of
// data, like slices/arrays and maps.
//
// For the functions that return an array, a []any is returned,
// regardless of whether or not the input was a different type.
package coll

import (
	"fmt"
	"maps"
	"math"
	"math/big"
	"reflect"
	"slices"
	"sort"

	"github.com/hairyhenderson/gomplate/v5/conv"
	iconv "github.com/hairyhenderson/gomplate/v5/internal/conv"
)

// Slice creates a slice from a bunch of arguments
func Slice(args ...any) []any {
	return args
}

// Has determines whether or not a given object has a property with the given key
func Has(in any, key any) bool {
	return iconv.Has(in, key)
}

// Dict is a convenience function that creates a map with string keys.
// Provide arguments as key/value pairs. If an odd number of arguments
// is provided, the last is used as the key, and an empty string is
// set as the value.
// All keys are converted to strings, regardless of input type.
func Dict(v ...any) (map[string]any, error) {
	dict := map[string]any{}
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
func Keys(in ...map[string]any) ([]string, error) {
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

func splitMap(m map[string]any) ([]string, []any) {
	keys := make([]string, len(m))
	values := make([]any, len(m))
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
func Values(in ...map[string]any) ([]any, error) {
	if len(in) == 0 {
		return nil, fmt.Errorf("need at least one argument")
	}
	values := []any{}
	for _, m := range in {
		_, v := splitMap(m)
		values = append(values, v...)
	}
	return values, nil
}

// Append v to the end of list. No matter what type of input slice or array list is, a new []any is always returned.
func Append(v any, list any) ([]any, error) {
	l, err := iconv.InterfaceSlice(list)
	if err != nil {
		return nil, err
	}

	return append(l, v), nil
}

// Prepend v to the beginning of list. No matter what type of input slice or array list is, a new []any is always returned.
func Prepend(v any, list any) ([]any, error) {
	l, err := iconv.InterfaceSlice(list)
	if err != nil {
		return nil, err
	}

	return append([]any{v}, l...), nil
}

// Uniq finds the unique values within list. No matter what type of input slice or array list is, a new []any is always returned.
func Uniq(list any) ([]any, error) {
	l, err := iconv.InterfaceSlice(list)
	if err != nil {
		return nil, err
	}

	out := make([]any, 0, len(l))

	// scalar values can be compared with ==, which gives the same result as
	// reflect.DeepEqual for them, so track those in a set. Everything else
	// (maps, slices, pointers, structs, etc) is compared with reflect.DeepEqual.
	// most lists are all scalars, so size the set for that, assuming about
	// half are duplicates
	seen := make(map[any]struct{}, len(l)/2)
	others := []any{}

	for _, v := range l {
		if isScalar(v) {
			if _, ok := seen[v]; ok {
				continue
			}

			seen[v] = struct{}{}
		} else {
			if slices.ContainsFunc(others, func(o any) bool {
				return reflect.DeepEqual(o, v)
			}) {
				continue
			}

			others = append(others, v)
		}

		out = append(out, v)
	}

	return out, nil
}

// isScalar returns true if v is nil, or a bool, number, or string.
func isScalar(v any) bool {
	switch v.(type) {
	case nil, string, int, int64, float64, bool:
		return true
	}

	switch reflect.TypeOf(v).Kind() {
	case reflect.Bool, reflect.String,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
		return true
	default:
		return false
	}
}

// Reverse the list. No matter what type of input slice or array list is, a new []any is always returned.
func Reverse(list any) ([]any, error) {
	l, err := iconv.InterfaceSlice(list)
	if err != nil {
		return nil, err
	}

	// InterfaceSlice returns the input as-is when it's already a []any, so
	// copy it to avoid reversing the caller's slice
	l = slices.Clone(l)

	// nifty trick from https://github.com/golang/go/wiki/SliceTricks#reversing
	for left, right := 0, len(l)-1; left < right; left, right = left+1, right-1 {
		l[left], l[right] = l[right], l[left]
	}
	return l, nil
}

// Merge source maps (srcs) into dst. Precedence is in left-to-right order, with
// the left-most values taking precedence over the right-most.
func Merge(dst map[string]any, srcs ...map[string]any) (map[string]any, error) {
	for _, src := range srcs {
		dst = mergeValues(src, dst)
	}
	return dst, nil
}

// Omit returns a new map without any entries that have the
// given keys (inverse of Pick).
func Omit(in map[string]any, keys ...string) map[string]any {
	out := map[string]any{}
	for k, v := range in {
		if !slices.Contains(keys, k) {
			out[k] = v
		}
	}
	return out
}

// OmitSlice returns a new slice without any elements equal to the given values.
// It always returns a new []any, regardless of the input slice or array type.
func OmitSlice(list any, values ...any) ([]any, error) {
	l, err := iconv.InterfaceSlice(list)
	if err != nil {
		return nil, err
	}

	out := []any{}
	for _, v := range l {
		if !slices.ContainsFunc(values, func(val any) bool {
			return reflect.DeepEqual(v, val)
		}) {
			out = append(out, v)
		}
	}
	return out, nil
}

// Pick returns a new map with any entries that have the
// given keys (inverse of Omit).
func Pick(in map[string]any, keys ...string) map[string]any {
	out := map[string]any{}
	for k, v := range in {
		if slices.Contains(keys, k) {
			out[k] = v
		}
	}
	return out
}

// Merges a default and override map
func mergeValues(d map[string]any, o map[string]any) map[string]any {
	def := maps.Clone(d)
	if def == nil {
		def = map[string]any{}
	}
	over := maps.Clone(o)
	for k, v := range over {
		// If the key doesn't exist already, then just set the key to that value
		if _, exists := def[k]; !exists {
			def[k] = v
			continue
		}
		nextMap, ok := v.(map[string]any)
		// If it isn't another map, overwrite the value
		if !ok {
			def[k] = v
			continue
		}
		// Edge case: If the key exists in the default, but isn't a map
		defMap, isMap := def[k].(map[string]any)
		// If the override map has a map for this key, prefer it
		if !isMap {
			def[k] = v
			continue
		}
		// If we got to this point, it is a map in both, so merge them
		def[k] = mergeValues(defMap, nextMap)
	}
	return def
}

// Sort a given array or slice. Uses natural sort order if possible. If a
// non-empty key is given and the list elements are maps, this will attempt to
// sort by the values of those entries.
//
// Does not modify the input list.
func Sort(key string, list any) (out []any, err error) {
	if list == nil {
		return nil, nil
	}

	ia, err := iconv.InterfaceSlice(list)
	if err != nil {
		return nil, err
	}
	// if the types are all the same, we can sort the slice
	if sameTypes(ia) {
		s := make([]any, len(ia))
		// make a copy so the original is unmodified
		copy(s, ia)
		sort.SliceStable(s, func(i, j int) bool {
			return lessThan(key)(s[i], s[j])
		})
		return s, nil
	}
	return ia, nil
}

// lessThan - compare two values. Values of different kinds are compared as
// numbers if they're both numbers, and otherwise by kind, so that the order
// is at least consistent.
func lessThan(key string) func(left, right any) bool {
	return func(left, right any) bool {
		val := reflect.Indirect(reflect.ValueOf(left))
		rval := reflect.Indirect(reflect.ValueOf(right))
		if val.Kind() != rval.Kind() {
			lf, lok := toBigFloat(val)
			rf, rok := toBigFloat(rval)
			if lok && rok {
				// NaN isn't less or greater than anything
				return lf != nil && rf != nil && lf.Cmp(rf) < 0
			}
			return val.Kind() < rval.Kind()
		}
		switch val.Kind() {
		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
			return val.Int() < rval.Int()
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint, reflect.Uint64, reflect.Uintptr:
			return val.Uint() < rval.Uint()
		case reflect.Float32, reflect.Float64:
			return val.Float() < rval.Float()
		case reflect.String:
			return val.String() < rval.String()
		case reflect.Map:
			return keyedLess(mapIndex(val, key), mapIndex(rval, key))
		case reflect.Struct:
			return keyedLess(val.FieldByName(key), rval.FieldByName(key))
		default:
			// it's not really comparable, so...
			return false
		}
	}
}

// mapIndex returns the value of key in m, or an invalid value if it's missing
// or m's keys can't be strings.
func mapIndex(m reflect.Value, key string) reflect.Value {
	kt := m.Type().Key()
	kval := reflect.ValueOf(key)
	switch {
	case kt.Kind() == reflect.String:
		return m.MapIndex(kval.Convert(kt))
	case kval.Type().AssignableTo(kt):
		return m.MapIndex(kval)
	default:
		return reflect.Value{}
	}
}

// keyedLess compares two values found by a map key or field name. Missing
// values sort after everything else.
func keyedLess(left, right reflect.Value) bool {
	if !left.IsValid() || !left.CanInterface() {
		return false
	}
	if !right.IsValid() || !right.CanInterface() {
		return true
	}
	return lessThan("")(left.Interface(), right.Interface())
}

// toBigFloat returns v as an exact *big.Float, if it's a number. NaN can't be
// represented, so it's returned as nil.
func toBigFloat(v reflect.Value) (*big.Float, bool) {
	switch v.Kind() {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		return new(big.Float).SetInt64(v.Int()), true
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint, reflect.Uint64, reflect.Uintptr:
		return new(big.Float).SetUint64(v.Uint()), true
	case reflect.Float32, reflect.Float64:
		if math.IsNaN(v.Float()) {
			return nil, true
		}
		return new(big.Float).SetFloat64(v.Float()), true
	default:
		return nil, false
	}
}

func sameTypes(a []any) bool {
	if len(a) == 0 {
		return true
	}
	t := reflect.TypeOf(a[0])
	for _, v := range a[1:] {
		vt := reflect.TypeOf(v)
		// a nil value has no type, so only matches another nil
		if t == nil || vt == nil {
			if t != vt {
				return false
			}
			continue
		}
		if vt.Kind() != t.Kind() {
			return false
		}
	}
	return true
}

// Flatten a nested array or slice to at most 'depth' levels. Use depth of -1
// to completely flatten the input.
// Returns a new slice without modifying the input.
func Flatten(list any, depth int) ([]any, error) {
	l, err := iconv.InterfaceSlice(list)
	if err != nil {
		return nil, err
	}
	if depth == 0 {
		return l, nil
	}
	out := make([]any, 0, len(l)*2)
	for _, v := range l {
		s := reflect.ValueOf(v)
		kind := s.Kind()
		switch kind {
		case reflect.Slice, reflect.Array:
			vl, err := Flatten(v, depth-1)
			if err != nil {
				return nil, err
			}
			out = append(out, vl...)
		default:
			out = append(out, v)
		}
	}
	return out, nil
}
