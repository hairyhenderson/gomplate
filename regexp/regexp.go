// Package regexp contains functions for dealing with regular expressions
package regexp

import stdre "regexp"

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
func Match(expression, input string) bool {
	re := stdre.MustCompile(expression)
	return re.MatchString(input)
}

// Replace -
func Replace(expression, replacement, input string) string {
	re := stdre.MustCompile(expression)
	return re.ReplaceAllString(input, replacement)
}

// ReplaceLiteral -
func ReplaceLiteral(expression, replacement, input string) (string, error) {
	re, err := stdre.Compile(expression)
	if err != nil {
		return "", err
	}
	return re.ReplaceAllLiteralString(input, replacement), nil
}

// Split -
func Split(expression string, n int, input string) ([]string, error) {
	re, err := stdre.Compile(expression)
	if err != nil {
		return nil, err
	}
	return re.Split(input, n), nil
}
