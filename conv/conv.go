package conv

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// Bool converts a string to a boolean value, using strconv.ParseBool under the covers.
// Possible true values are: 1, t, T, TRUE, true, True
// All other values are considered false.
//
// See ToBool also for a more flexible version.
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
func ToBool(in interface{}) bool {
	if b, ok := in.(bool); ok {
		return b
	}

	if str, ok := in.(string); ok {
		str = strings.ToLower(str)
		switch str {
		case "1", "t", "true", "yes":
			return true
		default:
			return strToFloat64(str) == 1
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
func ToBools(in ...interface{}) []bool {
	out := make([]bool, len(in))
	for i, v := range in {
		out[i] = ToBool(v)
	}
	return out
}

// Slice creates a slice from a bunch of arguments
func Slice(args ...interface{}) []interface{} {
	return args
}

// Join concatenates the elements of a to create a single string.
// The separator string sep is placed between elements in the resulting string.
//
// This is functionally identical to strings.Join, except that each element is
// coerced to a string first
func Join(in interface{}, sep string) (out string, err error) {
	s, ok := in.([]string)
	if ok {
		return strings.Join(s, sep), nil
	}

	var a []interface{}
	a, ok = in.([]interface{})
	if !ok {
		a, err = interfaceSlice(in)
		if err != nil {
			return "", errors.Wrap(err, "Input to Join must be an array")
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

	return "", errors.New("Input to Join must be an array")
}

func interfaceSlice(slice interface{}) ([]interface{}, error) {
	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice {
		return nil, errors.Errorf("interfaceSlice given a non-slice type %T", s)
	}
	ret := make([]interface{}, s.Len())
	for i := 0; i < s.Len(); i++ {
		ret[i] = s.Index(i).Interface()
	}
	return ret, nil
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

// ToString -
func ToString(in interface{}) string {
	if in == nil {
		return "nil"
	}
	if s, ok := in.(string); ok {
		return s
	}
	if s, ok := in.(fmt.Stringer); ok {
		return s.String()
	}

	v, ok := printableValue(reflect.ValueOf(in))
	if ok {
		in = v
	}
	return fmt.Sprint(in)
}

// ToStrings -
func ToStrings(in ...interface{}) []string {
	out := make([]string, len(in))
	for i, v := range in {
		out[i] = ToString(v)
	}
	return out
}

// MustParseInt - wrapper for strconv.ParseInt that returns 0 in the case of error
func MustParseInt(s string, base, bitSize int) int64 {
	// nolint: gosec
	i, _ := strconv.ParseInt(s, base, bitSize)
	return i
}

// MustParseFloat - wrapper for strconv.ParseFloat that returns 0 in the case of error
func MustParseFloat(s string, bitSize int) float64 {
	// nolint: gosec
	i, _ := strconv.ParseFloat(s, bitSize)
	return i
}

// MustParseUint - wrapper for strconv.ParseUint that returns 0 in the case of error
func MustParseUint(s string, base, bitSize int) uint64 {
	// nolint: gosec
	i, _ := strconv.ParseUint(s, base, bitSize)
	return i
}

// MustAtoi - wrapper for strconv.Atoi that returns 0 in the case of error
func MustAtoi(s string) int {
	// nolint: gosec
	i, _ := strconv.Atoi(s)
	return i
}

// ToInt64 - convert input to an int64, if convertible. Otherwise, returns 0.
func ToInt64(v interface{}) int64 {
	if str, ok := v.(string); ok {
		return strToInt64(str)
	}

	val := reflect.Indirect(reflect.ValueOf(v))
	switch val.Kind() {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		return val.Int()
	case reflect.Uint8, reflect.Uint16, reflect.Uint32:
		return int64(val.Uint())
	case reflect.Uint, reflect.Uint64:
		tv := val.Uint()
		// this can overflow and give -1, but IMO this is better than
		// returning maxint64
		return int64(tv)
	case reflect.Float32, reflect.Float64:
		return int64(val.Float())
	case reflect.Bool:
		if val.Bool() {
			return 1
		}
		return 0
	default:
		return 0
	}
}

// ToInt -
func ToInt(in interface{}) int {
	return int(ToInt64(in))
}

// ToInt64s -
func ToInt64s(in ...interface{}) []int64 {
	out := make([]int64, len(in))
	for i, v := range in {
		out[i] = ToInt64(v)
	}
	return out
}

// ToInts -
func ToInts(in ...interface{}) []int {
	out := make([]int, len(in))
	for i, v := range in {
		out[i] = ToInt(v)
	}
	return out
}

// ToFloat64 - convert input to a float64, if convertible. Otherwise, returns 0.
func ToFloat64(v interface{}) float64 {
	if str, ok := v.(string); ok {
		return strToFloat64(str)
	}

	val := reflect.Indirect(reflect.ValueOf(v))
	switch val.Kind() {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		return float64(val.Int())
	case reflect.Uint8, reflect.Uint16, reflect.Uint32:
		return float64(val.Uint())
	case reflect.Uint, reflect.Uint64:
		return float64(val.Uint())
	case reflect.Float32, reflect.Float64:
		return val.Float()
	case reflect.Bool:
		if val.Bool() {
			return 1
		}
		return 0
	default:
		return 0
	}
}

func strToInt64(str string) int64 {
	if strings.Contains(str, ",") {
		str = strings.Replace(str, ",", "", -1)
	}
	iv, err := strconv.ParseInt(str, 0, 64)
	if err != nil {
		// maybe it's a float?
		var fv float64
		fv, err = strconv.ParseFloat(str, 64)
		if err != nil {
			return 0
		}
		return ToInt64(fv)
	}
	return iv
}

func strToFloat64(str string) float64 {
	if strings.Contains(str, ",") {
		str = strings.Replace(str, ",", "", -1)
	}
	// this is inefficient, but it's the only way I can think of to
	// properly convert octal integers to floats
	iv, err := strconv.ParseInt(str, 0, 64)
	if err != nil {
		// ok maybe it's a float?
		var fv float64
		fv, err = strconv.ParseFloat(str, 64)
		if err != nil {
			return 0
		}
		return fv
	}
	return float64(iv)
}

// ToFloat64s -
func ToFloat64s(in ...interface{}) []float64 {
	out := make([]float64, len(in))
	for i, v := range in {
		out[i] = ToFloat64(v)
	}
	return out
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
		key := ToString(v[i])
		if i+1 >= lenv {
			dict[key] = ""
			continue
		}
		dict[key] = v[i+1]
	}
	return dict, nil
}
