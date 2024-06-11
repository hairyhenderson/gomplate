// Package regexp contains functions for dealing with regular expressions
package regexp

import (
	"fmt"
	stdre "regexp"
)

// Find -
func Find(expression, input string) (string, error) {
	re, err := stdre.Compile(expression)
	if err != nil {
		return "", err
	}
	return re.FindString(input), nil
}

// FindAll -
func FindAll(expression string, n int, input string) ([]string, error) {
	re, err := stdre.Compile(expression)
	if err != nil {
		return nil, err
	}
	return re.FindAllString(input, n), nil
}

// Match -
func Match(expression, input string) (bool, error) {
	re, err := stdre.Compile(expression)
	if err != nil {
		return false, fmt.Errorf("error compiling expression: %w", err)
	}

	return re.MatchString(input), nil
}

// QuoteMeta -
func QuoteMeta(input string) string {
	return stdre.QuoteMeta(input)
}

// Replace -
func Replace(expression, replacement, input string) (string, error) {
	re, err := stdre.Compile(expression)
	if err != nil {
		return "", fmt.Errorf("error compiling expression: %w", err)
	}

	return re.ReplaceAllString(input, replacement), nil
}

// ReplaceLiteral -
func ReplaceLiteral(expression, replacement, input string) (string, error) {
	re, err := stdre.Compile(expression)
	if err != nil {
		return "", fmt.Errorf("error compiling expression: %w", err)
	}
	return re.ReplaceAllLiteralString(input, replacement), nil
}

// Split -
func Split(expression string, n int, input string) ([]string, error) {
	re, err := stdre.Compile(expression)
	if err != nil {
		return nil, fmt.Errorf("error compiling expression: %w", err)
	}

	return re.Split(input, n), nil
}
