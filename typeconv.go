package main

import "strconv"

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
