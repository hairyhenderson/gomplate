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
