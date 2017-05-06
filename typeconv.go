package main

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"

	yaml "gopkg.in/yaml.v2"
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

func unmarshalObj(obj map[string]interface{}, in string, f func([]byte, interface{}) error) map[string]interface{} {
	err := f([]byte(in), &obj)
	if err != nil {
		log.Fatalf("Unable to unmarshal object %s: %v", in, err)
	}
	return obj
}

func unmarshalArray(obj []interface{}, in string, f func([]byte, interface{}) error) []interface{} {
	err := f([]byte(in), &obj)
	if err != nil {
		log.Fatalf("Unable to unmarshal array %s: %v", in, err)
	}
	return obj
}

// JSON - Unmarshal a JSON Object
func (t *TypeConv) JSON(in string) map[string]interface{} {
	obj := make(map[string]interface{})
	return unmarshalObj(obj, in, yaml.Unmarshal)
}

// JSONArray - Unmarshal a JSON Array
func (t *TypeConv) JSONArray(in string) []interface{} {
	obj := make([]interface{}, 1)
	return unmarshalArray(obj, in, yaml.Unmarshal)
}

// YAML - Unmarshal a YAML Object
func (t *TypeConv) YAML(in string) map[string]interface{} {
	obj := make(map[string]interface{})
	return unmarshalObj(obj, in, yaml.Unmarshal)
}

// YAMLArray - Unmarshal a YAML Array
func (t *TypeConv) YAMLArray(in string) []interface{} {
	obj := make([]interface{}, 1)
	return unmarshalArray(obj, in, yaml.Unmarshal)
}

func marshalObj(obj interface{}, f func(interface{}) ([]byte, error)) string {
	b, err := f(obj)
	if err != nil {
		log.Fatalf("Unable to marshal object %s: %v", obj, err)
	}

	return string(b)
}

// ToJSON - Stringify a struct as JSON
func (t *TypeConv) ToJSON(in interface{}) string {
	return marshalObj(in, json.Marshal)
}

// ToYAML - Stringify a struct as YAML
func (t *TypeConv) ToYAML(in interface{}) string {
	return marshalObj(in, yaml.Marshal)
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

// Has determines whether or not a given object has a property with the given key
func (t *TypeConv) Has(in interface{}, key string) bool {
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
