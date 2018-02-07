//+build integration
//+build !windows

package integration

import (
	. "gopkg.in/check.v1"

	"github.com/gotestyourself/gotestyourself/icmd"
)

type EnvvarsSuite struct{}

var _ = Suite(&EnvvarsSuite{})

func (s *EnvvarsSuite) TestNonExistantEnvVar(c *C) {
	result := icmd.RunCommand(GomplateBin, "-i",
		`{{ .Env.FOO }}`)
	result.Assert(c, icmd.Expected{ExitCode: 1, Err: "map has no entry for key"})

	result = icmd.RunCommand(GomplateBin, "-i",
		`{{ getenv "FOO" }}`)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: ""})

	result = icmd.RunCommand(GomplateBin, "-i",
		`{{ getenv "FOO" "foo" }}`)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "foo"})

	result = icmd.RunCmd(icmd.Command(GomplateBin, "-i", `{{ getenv "FOO" "foo" }}`),
		func(c *icmd.Cmd) {
			c.Env = []string{"FOO="}
		})
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "foo"})
}

func (s *EnvvarsSuite) TestExistantEnvVar(c *C) {
	setFoo := func(c *icmd.Cmd) {
		c.Env = []string{"FOO=foo"}
	}
	expected := icmd.Expected{ExitCode: 0, Out: "foo"}
	result := icmd.RunCmd(icmd.Command(GomplateBin, "-i",
		`{{ .Env.FOO }}`), setFoo)
	result.Assert(c, expected)

	result = icmd.RunCmd(icmd.Command(GomplateBin, "-i",
		`{{ getenv "FOO" }}`), setFoo)
	result.Assert(c, expected)

	result = icmd.RunCmd(icmd.Command(GomplateBin, "-i",
		`{{ env.Getenv "FOO" }}`), setFoo)
	result.Assert(c, expected)
}
