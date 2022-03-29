// Package strings contains functions to manipulate strings
package strings

import (
	"regexp"
	"sort"
	"strings"

	"github.com/Masterminds/goutils"
	"github.com/hairyhenderson/gomplate/v3/conv"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Indent - indent each line of the string with the given indent string.
// Any indent characters are permitted, except for '\n'.
//
// TODO: return an error if the indent string contains '\n' instead of
// succeeding
func Indent(width int, indent, s string) string {
	if width <= 0 || strings.Contains(indent, "\n") {
		return s
	}
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

// ShellQuote - generate a POSIX shell literal evaluating to a string
func ShellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
}

// Trunc - truncate a string to the given length
func Trunc(length int, s string) string {
	if length < 0 {
		return s
	}
	if len(s) <= length {
		return s
	}
	return s[0:length]
}

// Sort - return an alphanumerically-sorted list of strings
//
// Deprecated: use coll.Sort instead
func Sort(list []string) []string {
	sorted := sort.StringSlice(list)
	sorted.Sort()
	return sorted
}

var (
	spaces      = regexp.MustCompile(`\s+`)
	nonAlphaNum = regexp.MustCompile(`[^\pL\pN]+`)
)

// SnakeCase -
func SnakeCase(in string) string {
	s := casePrepare(in)
	return spaces.ReplaceAllString(s, "_")
}

// KebabCase -
func KebabCase(in string) string {
	s := casePrepare(in)
	return spaces.ReplaceAllString(s, "-")
}

func casePrepare(in string) string {
	in = strings.TrimSpace(in)
	s := strings.ToLower(in)
	// make sure the first letter remains lower- or upper-cased
	s = strings.Replace(s, string(s[0]), string(in[0]), 1)
	s = nonAlphaNum.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

// CamelCase -
func CamelCase(in string) string {
	in = strings.TrimSpace(in)
	tag := language.Und
	s := cases.Title(tag).String(conv.ToString(in))
	// make sure the first letter remains lower- or upper-cased
	s = strings.Replace(s, string(s[0]), string(in[0]), 1)
	return nonAlphaNum.ReplaceAllString(s, "")
}

// WordWrapOpts defines the options to apply to the WordWrap function
type WordWrapOpts struct {
	// Line-break sequence to insert (defaults to "\n")
	LBSeq string

	// The desired maximum line length in characters (defaults to 80)
	Width uint
}

// applies default options
func wwDefaults(opts WordWrapOpts) WordWrapOpts {
	if opts.Width == 0 {
		opts.Width = 80
	}
	if opts.LBSeq == "" {
		opts.LBSeq = "\n"
	}
	return opts
}

// WordWrap - insert line-breaks into the string, before it reaches the given
// width.
func WordWrap(in string, opts WordWrapOpts) string {
	opts = wwDefaults(opts)
	return goutils.WrapCustom(in, int(opts.Width), opts.LBSeq, false)
}
