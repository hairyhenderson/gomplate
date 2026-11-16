// Package strings contains functions to manipulate strings
package strings

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/Masterminds/goutils"
	"github.com/hairyhenderson/gomplate/v5/conv"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Indent - indent each line of the string with the given indent string.
// Any indent characters are permitted, except for '\n'.
func Indent(width int, indent, s string) (string, error) {
	if width <= 0 {
		return "", fmt.Errorf("width must be > 0")
	}

	if strings.Contains(indent, "\n") {
		return "", fmt.Errorf("indent must not contain '\\n'")
	}

	if width > 1 {
		indent = strings.Repeat(indent, width)
	}

	lines := strings.Count(s, "\n")

	res := make([]byte, 0, len(s)+len(indent)*lines)
	bol := true
	for i := range len(s) {
		c := s[i]
		if bol && c != '\n' {
			res = append(res, indent...)
		}
		res = append(res, c)
		bol = c == '\n'
	}

	return string(res), nil
}

// ShellQuote - generate a POSIX shell literal evaluating to a string
func ShellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
}

// Trunc - truncate a string to the given length (in bytes). A multi-byte
// character that doesn't fit is dropped instead of being cut in half.
func Trunc(length int, s string) string {
	if length < 0 {
		return s
	}
	if len(s) <= length {
		return s
	}

	// cut at the start of the last character that begins at or before length
	end := 0
	for i := range s {
		if i > length {
			break
		}
		end = i
	}

	return s[:end]
}

// Abbrev abbreviates s with "..." (ellipses) so that the result is at most
// width characters long. If offset is greater than 4, the result starts with
// "..." and shows the text from around offset. Widths and offsets count
// characters (runes), so multi-byte characters are never cut in half. For
// ASCII input this matches goutils.AbbreviateFull.
func Abbrev(offset, width int, s string) (string, error) {
	r := []rune(s)
	if len(r) <= width {
		return s, nil
	}

	if width < 4 {
		return "", fmt.Errorf("stringutils illegal argument: Minimum abbreviation width is 4")
	}

	const marker = "..."

	offset = min(offset, len(r))
	if len(r)-offset < width-3 {
		offset = len(r) - (width - 3)
	}

	if offset <= 4 {
		return string(r[:width-3]) + marker, nil
	}

	if width < 7 {
		return "", fmt.Errorf("stringutils illegal argument: Minimum abbreviation width with offset is 7")
	}

	if offset+width-3 < len(r) {
		rest, err := Abbrev(0, width-3, string(r[offset:]))
		if err != nil {
			return "", err
		}

		return marker + rest, nil
	}

	return marker + string(r[len(r)-(width-3):]), nil
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

// keepFirstRune makes sure the first letter of s remains lower- or upper-cased
// like the first letter of orig. conv is the conversion that produced s from
// orig. It may expand the first letter into several runes, so the whole
// expansion is replaced, not just one rune. Both must be non-empty.
func keepFirstRune(s, orig string, conv func(string) string) string {
	r, _ := utf8.DecodeRuneInString(orig)
	first := string(r)
	if prefix := conv(first); strings.HasPrefix(s, prefix) {
		return first + s[len(prefix):]
	}

	_, size := utf8.DecodeRuneInString(s)
	return first + s[size:]
}

func casePrepare(in string) string {
	in = strings.TrimSpace(in)
	if in == "" {
		return ""
	}

	s := keepFirstRune(strings.ToLower(in), in, strings.ToLower)
	s = nonAlphaNum.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

// CamelCase -
func CamelCase(in string) string {
	in = strings.TrimSpace(in)
	if in == "" {
		return ""
	}

	title := cases.Title(language.Und).String
	s := keepFirstRune(title(conv.ToString(in)), in, title)
	return nonAlphaNum.ReplaceAllString(s, "")
}

// WordWrapOpts defines the options to apply to the WordWrap function
type WordWrapOpts struct {
	// Line-break sequence to insert (defaults to "\n")
	LBSeq string

	// The desired maximum line length in characters (defaults to 80)
	Width uint32
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

	// if we're on a 32-bit system, we need to check for overflow. If the width
	// is greater than maxint, we'll just use maxint.
	width := int(opts.Width)
	if width == -1 {
		width = int(math.MaxInt)
	}

	return goutils.WrapCustom(in, width, opts.LBSeq, false)
}

// SkipLines - skip the given number of lines (ending with \n) from the string.
// If skip is greater than the number of lines in the string, an empty string is
// returned.
func SkipLines(skip int, in string) (string, error) {
	if skip < 0 {
		return "", fmt.Errorf("skip must be >= 0")
	}
	if skip == 0 {
		return in, nil
	}

	lines := strings.SplitN(in, "\n", skip+1)
	if skip >= len(lines) {
		return "", nil
	}

	return lines[skip], nil
}
