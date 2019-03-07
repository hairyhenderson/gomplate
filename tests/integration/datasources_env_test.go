//+build integration

package integration

import (
	"os"
	"runtime"

	. "gopkg.in/check.v1"

	"github.com/gotestyourself/gotestyourself/icmd"
)

type EnvDatasourcesSuite struct {
}

var _ = Suite(&EnvDatasourcesSuite{})

func (s *EnvDatasourcesSuite) SetUpSuite(c *C) {
	os.Setenv("HELLO_WORLD", "hello world")
	os.Setenv("HELLO_UNIVERSE", "hello universe")
	os.Setenv("FOO", "bar")
	os.Setenv("foo", "baz")
}

func (s *EnvDatasourcesSuite) TearDownSuite(c *C) {
	os.Unsetenv("HELLO_WORLD")
	os.Unsetenv("HELLO_UNIVERSE")
	os.Unsetenv("FOO")
	os.Unsetenv("foo")
}

func (s *EnvDatasourcesSuite) TestEnvDatasources(c *C) {
	result := icmd.RunCommand(GomplateBin,
		"-d", "foo=env:FOO",
		"-i", `{{ ds "foo" }}`,
	)
	// Windows envvars are case-insensitive
	if runtime.GOOS == "windows" {
		result.Assert(c, icmd.Expected{ExitCode: 0, Out: "baz"})
	} else {
		result.Assert(c, icmd.Expected{ExitCode: 0, Out: "bar"})
	}

	result = icmd.RunCommand(GomplateBin,
		"-d", "foo=env:///foo",
		"-i", `{{ ds "foo" }}`,
	)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "baz"})

	result = icmd.RunCmd(icmd.Command(GomplateBin,
		"-d", "e=env:json_value?type=application/json",
		"-i", `{{ (ds "e").value}}`,
	), func(c *icmd.Cmd) {
		c.Env = []string{
			`json_value={"value":"corge"}`,
		}
	})
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "corge"})
}
