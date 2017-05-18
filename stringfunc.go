package main

import "strings"

// stringFunc - string manipulation function wrappers
type stringFunc struct{}

func (t stringFunc) replaceAll(old, new, s string) string {
	return strings.Replace(s, old, new, -1)
}
