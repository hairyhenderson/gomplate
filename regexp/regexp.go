package regexp

import stdre "regexp"

// Replace -
func Replace(expression, replacement, input string) string {
	re := stdre.MustCompile(expression)
	return re.ReplaceAllString(input, replacement)
}

// Match -
func Match(expression, input string) bool {
	re := stdre.MustCompile(expression)
	return re.MatchString(input)
}
