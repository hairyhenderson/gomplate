//+build integration

package integration

import (
	"bytes"

	. "gopkg.in/check.v1"

	"gotest.tools/v3/icmd"
)

type TestSuite struct {
}

var _ = Suite(&TestSuite{})

func (s *TestSuite) SetUpTest(c *C) {
}

func (s *TestSuite) TearDownTest(c *C) {
}

func (s *TestSuite) TestFail(c *C) {
	result := icmd.RunCommand(GomplateBin, "-i", "{{ fail }}")
	result.Assert(c, icmd.Expected{ExitCode: 1, Err: `template generation failed`})

	result = icmd.RunCommand(GomplateBin, "-i", "{{ fail `some message` }}")
	result.Assert(c, icmd.Expected{ExitCode: 1, Err: `some message`})
}

func (s *TestSuite) TestRequired(c *C) {
	result := icmd.RunCmd(icmd.Command(GomplateBin,
		"-i", `{{getenv "FOO" | required "FOO missing" }}`))
	result.Assert(c, icmd.Expected{
		ExitCode: 1,
		Err:      "FOO missing",
	})

	result = icmd.RunCmd(icmd.Command(GomplateBin,
		"-i", `{{getenv "FOO" | required "FOO missing" }}`),
		func(c *icmd.Cmd) {
			c.Env = []string{"FOO=bar"}
		})
	result.Assert(c, icmd.Expected{
		ExitCode: 0,
		Out:      "bar",
	})

	result = icmd.RunCmd(icmd.Command(GomplateBin,
		"-d", "in=stdin:///?type=application/yaml",
		"-i", `{{ (ds "in").foo | required "foo should not be null" }}`),
		func(c *icmd.Cmd) {
			c.Stdin = bytes.NewBufferString(`foo: null`)
		})
	result.Assert(c, icmd.Expected{
		ExitCode: 1,
		Err:      "foo should not be null",
	})

	result = icmd.RunCmd(icmd.Command(GomplateBin,
		"-d", "in=stdin:///?type=application/yaml",
		"-i", `{{ (ds "in").foo | required }}`),
		func(c *icmd.Cmd) {
			c.Stdin = bytes.NewBufferString(`foo: []`)
		})
	result.Assert(c, icmd.Expected{
		ExitCode: 0,
		Out:      "[]",
	})

	result = icmd.RunCmd(icmd.Command(GomplateBin,
		"-d", "in=stdin:///?type=application/yaml",
		"-i", `{{ (ds "in").foo | required }}`),
		func(c *icmd.Cmd) {
			c.Stdin = bytes.NewBufferString(`foo: {}`)
		})
	result.Assert(c, icmd.Expected{
		ExitCode: 0,
		Out:      "map[]",
	})

	result = icmd.RunCmd(icmd.Command(GomplateBin,
		"-d", "in=stdin:///?type=application/yaml",
		"-i", `{{ (ds "in").foo | required }}`),
		func(c *icmd.Cmd) {
			c.Stdin = bytes.NewBufferString(`foo: 0`)
		})
	result.Assert(c, icmd.Expected{
		ExitCode: 0,
		Out:      "0",
	})

	result = icmd.RunCmd(icmd.Command(GomplateBin,
		"-d", "in=stdin:///?type=application/yaml",
		"-i", `{{ (ds "in").foo | required }}`),
		func(c *icmd.Cmd) {
			c.Stdin = bytes.NewBufferString(`foo: false`)
		})
	result.Assert(c, icmd.Expected{
		ExitCode: 0,
		Out:      "false",
	})
}
