package conv

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
)

// Bool converts a string to a boolean value, using strconv.ParseBool under the covers.
// Possible true values are: 1, t, T, TRUE, true, True
// All other values are considered false.
func Bool(in string) bool {
	if b, err := strconv.ParseBool(in); err == nil {
		return b
	}
	return false
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
func Join(in interface{}, sep string) string {
	s, ok := in.([]string)
	if ok {
		return strings.Join(s, sep)
	}

	var a []interface{}
	a, ok = in.([]interface{})
	if !ok {
		var err error
		a, err = interfaceSlice(in)
		ok = err == nil
	}
	if ok {
		b := make([]string, len(a))
		for i := range a {
			b[i] = toString(a[i])
		}
		return strings.Join(b, sep)
	}

	log.Fatal("Input to Join must be an array")
	return ""
}

func interfaceSlice(slice interface{}) ([]interface{}, error) {
	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice {
		return nil, fmt.Errorf("interfaceSlice given a non-slice type %T", s)
	}
	ret := make([]interface{}, s.Len())
	for i := 0; i < s.Len(); i++ {
		ret[i] = s.Index(i).Interface()
	}
	return ret, nil
}

// Has determines whether or not a given object has a property with the given key
func Has(in interface{}, key string) bool {
	av := reflect.ValueOf(in)
	kv := reflect.ValueOf(key)

	if av.Kind() == reflect.Map {
		return av.MapIndex(kv).IsValid()
	}

	return false
}

func toString(in interface{}) string {
	if in == nil {
		return "nil"
	}
	if s, ok := in.(string); ok {
		return s
	}
	if s, ok := in.(fmt.Stringer); ok {
		return s.String()
	}
	val := reflect.Indirect(reflect.ValueOf(in))
	switch val.Kind() {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		return strconv.FormatInt(val.Int(), 10)
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint, reflect.Uint64:
		return strconv.FormatUint(val.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(val.Float(), 'f', -1, 64)
	case reflect.Bool:
		return strconv.FormatBool(val.Bool())
	default:
		return fmt.Sprintf("%s", in)
	}
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

// ToInt64 - taken from github.com/Masterminds/sprig
func ToInt64(v interface{}) int64 {
	if str, ok := v.(string); ok {
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
		if val.Bool() == true {
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

// ToFloat64 - taken from github.com/Masterminds/sprig
func ToFloat64(v interface{}) float64 {
	if str, ok := v.(string); ok {
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
		if val.Bool() == true {
			return 1
		}
		return 0
	default:
		return 0
	}
}

// ToFloat64s -
func ToFloat64s(in ...interface{}) []float64 {
	out := make([]float64, len(in))
	for i, v := range in {
		out[i] = ToFloat64(v)
	}
	return out
}
