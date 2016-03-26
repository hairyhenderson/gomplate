package main

import (
	"encoding/json"
	"log"
	"strconv"
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

// JSON - Unmarshal a JSON string
func (t *TypeConv) JSON(in string) map[string]interface{} {
	obj := make(map[string]interface{})
	err := json.Unmarshal([]byte(in), &obj)
	if err != nil {
		log.Fatalf("Unable to unmarshal JSON object %s: %v", in, err)
	}
	return obj
}
