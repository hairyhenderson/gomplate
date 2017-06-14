package strings

import "strings"

// Indent - indent each line of the string with the given indent string
func Indent(width int, indent, s string) string {
	if width > 1 {
		indent = strings.Repeat(indent, width)
	}
	var res []byte
	bol := true
	for i := 0; i < len(s); i++ {
		c := s[i]
		if bol && c != '\n' {
			res = append(res, indent...)
		}
		res = append(res, c)
		bol = c == '\n'
	}
	return string(res)
}
