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
	if s, ok := in.(string); ok {
		return s
	}
	if s, ok := in.(fmt.Stringer); ok {
		return s.String()
	}
	if i, ok := in.(int); ok {
		return strconv.Itoa(i)
	}
	if u, ok := in.(uint64); ok {
		return strconv.FormatUint(u, 10)
	}
	if f, ok := in.(float64); ok {
		return strconv.FormatFloat(f, 'f', -1, 64)
	}
	if b, ok := in.(bool); ok {
		return strconv.FormatBool(b)
	}
	if in == nil {
		return "nil"
	}
	return fmt.Sprintf("%s", in)
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
