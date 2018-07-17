//+build integration
//+build !windows

package integration

import (
	. "gopkg.in/check.v1"

	"github.com/gotestyourself/gotestyourself/icmd"
)

type StringsSuite struct{}

var _ = Suite(&StringsSuite{})

func (s *StringsSuite) TestIndent(c *C) {
	result := icmd.RunCommand(GomplateBin, "-i",
		`{{ strings.Indent "   " "hello world" }}
{{ "hello\nmultiline\nworld" | indent 2 "-" }}
{{ "foo\nbar" | strings.Indent 2 }}
    {{"hello\nworld" | strings.Indent 5 | strings.TrimSpace }}
`)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: `   hello world
--hello
--multiline
--world
  foo
  bar
    hello
     world`})
}

func (s *StringsSuite) TestRepeat(c *C) {
	result := icmd.RunCommand(GomplateBin, "-i",
		`ba{{ strings.Repeat 2 "na" }}`)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: `banana`})

	result = icmd.RunCommand(GomplateBin, "-i",
		`ba{{ strings.Repeat 9223372036854775807 "na" }}`)
	result.Assert(c, icmd.Expected{ExitCode: 1, Out: `too long: causes overflow`})

	result = icmd.RunCommand(GomplateBin, "-i",
		`ba{{ strings.Repeat -1 "na" }}`)
	result.Assert(c, icmd.Expected{ExitCode: 1, Out: `negative count`})
}

func (s *StringsSuite) TestSlug(c *C) {
	result := icmd.RunCommand(GomplateBin, "-i",
		`{{ strings.Slug "Hellö, Wôrld! Free @ last..." }}`)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: `hello-world-free-at-last`})
}
