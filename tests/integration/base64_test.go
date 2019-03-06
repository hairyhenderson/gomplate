//+build integration

package integration

import (
	. "gopkg.in/check.v1"

	"github.com/gotestyourself/gotestyourself/icmd"
)

type Base64Suite struct{}

var _ = Suite(&Base64Suite{})

func (s *Base64Suite) TestBase64Encode(c *C) {
	result := icmd.RunCommand(GomplateBin, "-i",
		`{{ "foo" | base64.Encode }}`)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "Zm9v"})
}

func (s *Base64Suite) TestBase64Decode(c *C) {
	result := icmd.RunCommand(GomplateBin, "-i",
		`{{ "Zm9v" | base64.Decode }}`)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "foo"})
}
