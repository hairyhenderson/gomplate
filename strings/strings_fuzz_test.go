package strings

import (
	"strings"
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
)

func FuzzIndent(f *testing.F) {
	f.Add(0, "  ", "foo\n")
	f.Add(1, "  ", "bar\nbaz\nqux\n")
	f.Add(1, "  ", "\n")
	f.Add(3, "  ", "quux\n")
	f.Add(15, "\n0", "\n0")

	f.Fuzz(func(t *testing.T, width int, indent, s string) {
		out := Indent(width, indent, s)

		// out should be equal to s when both have the indent character
		// completely removed.
		assert.Equal(t,
			strings.ReplaceAll(s, indent, ""),
			strings.ReplaceAll(out, indent, ""),
		)
	})
}

func FuzzTrunc(f *testing.F) {
	f.Add(0, "foo")

	f.Fuzz(func(t *testing.T, length int, s string) {
		out := Trunc(length, s)

		assert.LessOrEqual(t, len(out), len(s))
		if length >= 0 {
			assert.LessOrEqual(t, len(out), length)
		}

		assert.Equal(t, s[0:len(out)], out)
	})
}

func FuzzWordWrap(f *testing.F) {
	out := `There shouldn't be any wrapping of long words or URLs because that would break
things very badly. To wit:
https://example.com/a/super-long/url/that-shouldnt-be?wrapped=for+fear+of#the-breaking-of-functionality
should appear on its own line, regardless of the desired word-wrapping width
that has been set.`
	f.Add(out, "", uint(0))
	f.Add(out, "\n", uint(80))
	f.Add(out, "\v", uint(10))

	f.Fuzz(func(t *testing.T, in, lbSeq string, width uint) {
		for _, r := range lbSeq {
			if !unicode.IsSpace(r) {
				t.Skip("ignore non-whitespace sequences")
			}
		}

		out := WordWrap(in, WordWrapOpts{
			LBSeq: lbSeq,
			Width: width,
		})

		if lbSeq == "" {
			lbSeq = "\n"
		}
		// compare by stripping both the line-break sequence and spaces
		assert.Equal(t,
			strings.ReplaceAll(strings.ReplaceAll(in, lbSeq, ""), " ", ""),
			strings.ReplaceAll(strings.ReplaceAll(out, lbSeq, ""), " ", ""),
		)
	})
}
