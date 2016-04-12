package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
)

// TypeConv - type conversion function
type TypeConv struct {
}

// Bool converts a string to a boolean value, using strconv.ParseBool under the covers.
// Possible true values are: 1, t, T, TRUE, true, True
// All other values are considered false.
func (t *TypeConv) Bool(in string) bool {
	if b, err := strconv.ParseBool(in); err == nil {
		return b
	}
	return false
}

// JSON - Unmarshal a JSON Object
func (t *TypeConv) JSON(in string) map[string]interface{} {
	obj := make(map[string]interface{})
	err := json.Unmarshal([]byte(in), &obj)
	if err != nil {
		log.Fatalf("Unable to unmarshal JSON object %s: %v", in, err)
	}
	return obj
}

// JSONArray - Unmarshal a JSON Array
func (t *TypeConv) JSONArray(in string) []interface{} {
	obj := make([]interface{}, 1)
	err := json.Unmarshal([]byte(in), &obj)
	if err != nil {
		log.Fatalf("Unable to unmarshal JSON array %s: %v", in, err)
	}
	return obj
}

// Slice creates a slice from a bunch of arguments
func (t *TypeConv) Slice(args ...interface{}) []interface{} {
	return args
}

// Join concatenates the elements of a to create a single string.
// The separator string sep is placed between elements in the resulting string.
//
// This is functionally identical to strings.Join, except that each element is
// coerced to a string first
func (t *TypeConv) Join(a []interface{}, sep string) string {
	b := make([]string, len(a))
	for i := range a {
		b[i] = toString(a[i])
	}
	return strings.Join(b, sep)
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
