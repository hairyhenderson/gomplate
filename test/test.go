package test

import (
	"github.com/pkg/errors"
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
