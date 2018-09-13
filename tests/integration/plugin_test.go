//+build integration
//+build !windows

package integration

import (
	. "gopkg.in/check.v1"
	"github.com/gotestyourself/gotestyourself/icmd"
)

type PluginSuite struct {
}

var _ = Suite(&PluginSuite{})

func (s *EnvDatasourcesSuite) TestPlugin(c *C) {
	//make file compiles the following plugin file
	result := icmd.RunCommand(GomplateBin,
	"-p", "\"../testPlugin.so\"",
		"-d", "foo=testname:../..",
		"-i", "\'{{ ds \"foo\" }}\'",
	)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "bar"})
}
