// Package conv contains functions that help converting data between different types
package conv

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"

	iconv "github.com/hairyhenderson/gomplate/v4/internal/conv"
)

// Bool converts a string to a boolean value, using strconv.ParseBool under the covers.
// Possible true values are: 1, t, T, TRUE, true, True
// All other values are considered false.
//
// See ToBool also for a more flexible version.
//
// Deprecated: use ToBool instead
func Bool(in string) bool {
	if b, err := strconv.ParseBool(in); err == nil {
		return b
	}
	return false
}

// ToBool converts an arbitrary input into a boolean.
// Possible non-boolean true values are: 1 or the strings "t", "true", or "yes"
// (any capitalizations)
// All other values are considered false.
func ToBool(in any) bool {
	if b, ok := in.(bool); ok {
		return b
	}

	if str, ok := in.(string); ok {
		str = strings.ToLower(str)
		switch str {
		case "1", "t", "true", "yes":
			return true
		default:
			// ignore error here, as we'll just return false
			f, _ := strToFloat64(str)
			return f == 1
		}
	}

	val := reflect.Indirect(reflect.ValueOf(in))
	switch val.Kind() {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		return val.Int() == 1
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		return val.Uint() == 1
	case reflect.Float32, reflect.Float64:
		return val.Float() == 1
	default:
		return false
	}
}

// ToBools -
func ToBools(in ...any) []bool {
	out := make([]bool, len(in))
	for i, v := range in {
		out[i] = ToBool(v)
	}
	return out
}

// Slice creates a slice from a bunch of arguments
//
// Deprecated: use [github.com/hairyhenderson/gomplate/v4/coll.Slice] instead
func Slice(args ...any) []any {
	return args
}

// Join concatenates the elements of a to create a single string.
// The separator string sep is placed between elements in the resulting string.
//
// This is functionally identical to strings.Join, except that each element is
// coerced to a string first
func Join(in any, sep string) (out string, err error) {
	s, ok := in.([]string)
	if ok {
		return strings.Join(s, sep), nil
	}

	var a []any
	a, ok = in.([]any)
	if !ok {
		a, err = iconv.InterfaceSlice(in)
		if err != nil {
			return "", fmt.Errorf("input to Join must be an array: %w", err)
		}
		ok = true
	}
	if ok {
		b := make([]string, len(a))
		for i := range a {
			b[i] = ToString(a[i])
		}
		return strings.Join(b, sep), nil
	}

	return "", fmt.Errorf("input to Join must be an array")
}

// Has determines whether or not a given object has a property with the given key
func Has(in any, key any) bool {
	av := reflect.ValueOf(in)

	switch av.Kind() {
	case reflect.Map:
		kv := reflect.ValueOf(key)
		return av.MapIndex(kv).IsValid()
	case reflect.Slice, reflect.Array:
		l := av.Len()
		for i := range l {
			v := av.Index(i).Interface()
			if reflect.DeepEqual(v, key) {
				return true
			}
		}
	}

	return false
}

// ToString -
func ToString(in any) string {
	if in == nil {
		return "nil"
	}
	if s, ok := in.(string); ok {
		return s
	}
	if s, ok := in.(fmt.Stringer); ok {
		return s.String()
	}
	if s, ok := in.([]byte); ok {
		return string(s)
	}

	v, ok := printableValue(reflect.ValueOf(in))
	if ok {
		in = v
	}

	return fmt.Sprint(in)
}

// ToStrings -
func ToStrings(in ...any) []string {
	out := make([]string, len(in))
	for i, v := range in {
		out[i] = ToString(v)
	}
	return out
}

// MustParseInt - wrapper for strconv.ParseInt that returns 0 in the case of error
func MustParseInt(s string, base, bitSize int) int64 {
	i, _ := strconv.ParseInt(s, base, bitSize)
	return i
}

// MustParseFloat - wrapper for strconv.ParseFloat that returns 0 in the case of error
func MustParseFloat(s string, bitSize int) float64 {
	i, _ := strconv.ParseFloat(s, bitSize)
	return i
}

// MustParseUint - wrapper for strconv.ParseUint that returns 0 in the case of error
func MustParseUint(s string, base, bitSize int) uint64 {
	i, _ := strconv.ParseUint(s, base, bitSize)
	return i
}

// MustAtoi - wrapper for strconv.Atoi that returns 0 in the case of error
func MustAtoi(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

// ToInt64 - convert input to an int64, if convertible. Otherwise, returns 0.
func ToInt64(v any) (int64, error) {
	if str, ok := v.(string); ok {
		return strToInt64(str)
	}

	val := reflect.Indirect(reflect.ValueOf(v))
	switch val.Kind() {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		return val.Int(), nil
	case reflect.Uint8, reflect.Uint16, reflect.Uint32:
		//nolint:gosec // G115 isn't applicable, this is a Uint32 at most
		return int64(val.Uint()), nil
	case reflect.Uint, reflect.Uint64:
		tv := val.Uint()

		if tv > math.MaxInt64 {
			return 0, fmt.Errorf("could not convert %d to int64, would overflow", tv)
		}

		return int64(tv), nil
	case reflect.Float32, reflect.Float64:
		return int64(val.Float()), nil
	case reflect.Bool:
		if val.Bool() {
			return 1, nil
		}

		return 0, nil
	default:
		return 0, fmt.Errorf("could not convert %v to int64", v)
	}
}

// ToInt -
func ToInt(in any) (int, error) {
	i, err := ToInt64(in)
	if err != nil {
		return 0, err
	}

	// Bounds-checking to protect against CWE-190 and CWE-681
	// https://cwe.mitre.org/data/definitions/190.html
	// https://cwe.mitre.org/data/definitions/681.html
	if i >= math.MinInt && i <= math.MaxInt {
		return int(i), nil
	}

	// maybe we're on a 32-bit system, so we can't represent this number
	return 0, fmt.Errorf("could not convert %v to int", in)
}

// ToInt64s -
func ToInt64s(in ...any) ([]int64, error) {
	out := make([]int64, len(in))
	for i, v := range in {
		n, err := ToInt64(v)
		if err != nil {
			return nil, err
		}

		out[i] = n
	}

	return out, nil
}

// ToInts -
func ToInts(in ...any) ([]int, error) {
	out := make([]int, len(in))
	for i, v := range in {
		n, err := ToInt(v)
		if err != nil {
			return nil, err
		}

		out[i] = n
	}

	return out, nil
}

// ToFloat64 - convert input to a float64, if convertible. Otherwise, errors.
func ToFloat64(v any) (float64, error) {
	if str, ok := v.(string); ok {
		return strToFloat64(str)
	}

	val := reflect.Indirect(reflect.ValueOf(v))
	switch val.Kind() {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		return float64(val.Int()), nil
	case reflect.Uint8, reflect.Uint16, reflect.Uint32:
		return float64(val.Uint()), nil
	case reflect.Uint, reflect.Uint64:
		return float64(val.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return val.Float(), nil
	case reflect.Bool:
		if val.Bool() {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("could not convert %v to float64", v)
	}
}

func strToInt64(str string) (int64, error) {
	if strings.Contains(str, ",") {
		str = strings.ReplaceAll(str, ",", "")
	}

	iv, err := strconv.ParseInt(str, 0, 64)
	if err != nil {
		// maybe it's a float?
		var fv float64
		fv, err = strconv.ParseFloat(str, 64)
		if err != nil {
			return 0, fmt.Errorf("could not convert %q to int64: %w", str, err)
		}

		return ToInt64(fv)
	}

	return iv, nil
}

func strToFloat64(str string) (float64, error) {
	if strings.Contains(str, ",") {
		str = strings.ReplaceAll(str, ",", "")
	}

	// this is inefficient, but it's the only way I can think of to
	// properly convert octal integers to floats
	iv, err := strconv.ParseInt(str, 0, 64)
	if err != nil {
		// ok maybe it's a float?
		var fv float64
		fv, err = strconv.ParseFloat(str, 64)
		if err != nil {
			return 0, fmt.Errorf("could not convert %q to float64: %w", str, err)
		}

		return fv, nil
	}

	return float64(iv), nil
}

// ToFloat64s -
func ToFloat64s(in ...any) ([]float64, error) {
	out := make([]float64, len(in))
	for i, v := range in {
		f, err := ToFloat64(v)
		if err != nil {
			return nil, err
		}
		out[i] = f
	}

	return out, nil
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
		key := ToString(v[i])
		if i+1 >= lenv {
			dict[key] = ""
			continue
		}

		dict[key] = v[i+1]
	}

	return dict, nil
}
