package test

import (
	"github.com/pkg/errors"
	// "reflect"
)

// Assert -
func Assert(value bool, message string) (string, error) {
	if !value {
		if message != "" {
			return "", errors.Errorf("assertion failed: %s", message)
		}
		return "", errors.New("assertion failed")
	}
	return "", nil
}

// Fail -
func Fail(message string) error {
	if message != "" {
		return errors.Errorf("template generation failed: %s", message)
	}
	return errors.New("template generation failed")
}

// Required -
func Required(message string, value interface{}) (interface{}, error) {
	if message == "" {
		message = "can not render template: a required value was not set"
	}

	if s, ok := value.(string); value == nil || (ok && s == "") {
		return nil, errors.New(message)
	}

	return value, nil
}
