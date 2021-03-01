package integration

import (
	"strings"

	. "gopkg.in/check.v1"

	"gotest.tools/v3/assert"
)

type StringsSuite struct{}

var _ = Suite(&StringsSuite{})

func (s *StringsSuite) TestIndent(c *C) {
	inOutTest(c, `{{ strings.Indent "   " "hello world" }}
{{ "hello\nmultiline\nworld" | indent 2 "-" }}
{{ "foo\nbar" | strings.Indent 2 }}
    {{"hello\nworld" | strings.Indent 5 | strings.TrimSpace }}
`, `   hello world
--hello
--multiline
--world
  foo
  bar
    hello
     world
`)
}

func (s *StringsSuite) TestRepeat(c *C) {
	inOutTest(c, `ba{{ strings.Repeat 2 "na" }}`, `banana`)

	_, _, err := cmdTest(c, "-i", `ba{{ strings.Repeat 9223372036854775807 "na" }}`)
	assert.ErrorContains(c, err, `too long: causes overflow`)

	_, _, err = cmdTest(c, "-i", `ba{{ strings.Repeat -1 "na" }}`)
	assert.ErrorContains(c, err, `negative count`)
}

func (s *StringsSuite) TestSlug(c *C) {
	inOutTest(c, `{{ strings.Slug "Hellö, Wôrld! Free @ last..." }}`, `hello-world-free-at-last`)
}

func (s *StringsSuite) TestCaseFuncs(c *C) {
	inOutTest(c, `{{ strings.CamelCase "Hellö, Wôrld! Free @ last..." }}
{{ strings.SnakeCase "Hellö, Wôrld! Free @ last..." }}
{{ strings.KebabCase "Hellö, Wôrld! Free @ last..." }}`, `HellöWôrldFreeLast
Hellö_wôrld_free_last
Hellö-wôrld-free-last`)
}

func (s *StringsSuite) TestWordWrap(c *C) {
	out := `There shouldn't be any wrapping of long words or URLs because that would break
things very badly. To wit:
https://example.com/a/super-long/url/that-shouldnt-be?wrapped=for+fear+of#the-breaking-of-functionality
should appear on its own line, regardless of the desired word-wrapping width
that has been set.`
	text := strings.ReplaceAll(out, "\n", " ")
	in := `{{ print "` + text + `" | strings.WordWrap 80 }}`
	inOutTest(c, in, out)
}
