// Package test contains functions to help validate assumptions and can cause
// template generation to fail in specific cases
package test

import (
	"fmt"
)

// Assert -
func Assert(value bool, message string) (string, error) {
	if !value {
		if message != "" {
			return "", fmt.Errorf("assertion failed: %s", message)
		}
		return "", fmt.Errorf("assertion failed")
	}
	return "", nil
}

// Fail -
func Fail(message string) error {
	if message != "" {
		return fmt.Errorf("template generation failed: %s", message)
	}
	return fmt.Errorf("template generation failed")
}

// Required -
func Required(message string, value interface{}) (interface{}, error) {
	if message == "" {
		message = "can not render template: a required value was not set"
	}

	if s, ok := value.(string); value == nil || (ok && s == "") {
		return nil, fmt.Errorf(message)
	}

	return value, nil
}
