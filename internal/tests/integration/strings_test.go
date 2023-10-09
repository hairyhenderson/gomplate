package integration

import (
	"strings"
	"testing"

	"gotest.tools/v3/assert"
)

func TestStrings_Indent(t *testing.T) {
	inOutTest(t, `{{ strings.Indent "   " "hello world" }}
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

func TestStrings_Repeat(t *testing.T) {
	inOutTest(t, `ba{{ strings.Repeat 2 "na" }}`, `banana`)

	_, _, err := cmd(t, "-i", `ba{{ strings.Repeat 9223372036854775807 "na" }}`).run()
	assert.ErrorContains(t, err, `too long: causes overflow`)

	_, _, err = cmd(t, "-i", `ba{{ strings.Repeat -1 "na" }}`).run()
	assert.ErrorContains(t, err, `negative count`)
}

func TestStrings_Join(t *testing.T) {
	inOutTest(t, `{{ strings.Join (coll.Slice "hello" "world") " " }}`, `hello world`)
	inOutTest(t, `{{ strings.Join (coll.Slice 1 2 3) "," }}`, `1,2,3`)
}

func TestStrings_Slug(t *testing.T) {
	inOutTest(t, `{{ strings.Slug "Hellö, Wôrld! Free @ last..." }}`, `hello-world-free-at-last`)
}

func TestStrings_CaseFuncs(t *testing.T) {
	inOutTest(t, `{{ strings.ToLower "HELLO" }}
{{ strings.ToUpper "hello" }}
{{ strings.Title "is there anybody out there?" }}
{{ strings.Title "foo,bar᳇ǆaz"}}
`,
		`hello
HELLO
Is There Anybody Out There?
Foo,Bar᳇ǅaz
`)
	inOutTest(t, `{{ strings.CamelCase "Hellö, Wôrld! Free @ last..." }}
{{ strings.SnakeCase "Hellö, Wôrld! Free @ last..." }}
{{ strings.KebabCase "Hellö, Wôrld! Free @ last..." }}`, `HellöWôrldFreeLast
Hellö_wôrld_free_last
Hellö-wôrld-free-last`)
}

func TestStrings_WordWrap(t *testing.T) {
	out := `There shouldn't be any wrapping of long words or URLs because that would break
things very badly. To wit:
https://example.com/a/super-long/url/that-shouldnt-be?wrapped=for+fear+of#the-breaking-of-functionality
should appear on its own line, regardless of the desired word-wrapping width
that has been set.`
	text := strings.ReplaceAll(out, "\n", " ")
	in := `{{ print "` + text + `" | strings.WordWrap 80 }}`
	inOutTest(t, in, out)
}
