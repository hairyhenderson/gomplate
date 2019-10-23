//+build integration
//+build windows

package integration

import (
	. "gopkg.in/check.v1"

	"gotest.tools/v3/fs"
	"gotest.tools/v3/icmd"
)

type PluginsSuite struct {
	tmpDir *fs.Dir
}

var _ = Suite(&PluginsSuite{})

func (s *PluginsSuite) SetUpSuite(c *C) {
	s.tmpDir = fs.NewDir(c, "gomplate-inttests",
		fs.WithFile("foo.ps1", "echo $args\r\nexit 0\r\n", fs.WithMode(0644)),
		fs.WithFile("foo.bat", "@ECHO OFF\r\nECHO %1\r\n", fs.WithMode(0644)),
		fs.WithFile("fail.bat", `@ECHO OFF
ECHO %1 1>&2
EXIT /B %2
`, fs.WithMode(0755)),
		fs.WithFile("fail.ps1", `param (
       [Parameter(Position=0)]
       [string]$msg,

       [Parameter(Position=1)]
       [int]$code
)
$host.ui.WriteErrorLine($msg)
$host.SetShouldExit($code)
`, fs.WithMode(0755)),
	)
}

func (s *PluginsSuite) TearDownSuite(c *C) {
	s.tmpDir.Remove()
}

func (s *PluginsSuite) TestPlugins(c *C) {
	result := icmd.RunCommand(GomplateBin,
		"--plugin", "foo="+s.tmpDir.Join("foo.bat"),
		"-i", `{{ foo "hello world" }}`,
	)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "hello world"})
}

func (s *PluginsSuite) TestPluginErrors(c *C) {
	result := icmd.RunCmd(icmd.Command(GomplateBin,
		"--plugin", "f=false",
		"-i", `{{ f }}`,
	), func(c *icmd.Cmd) {})
	result.Assert(c, icmd.Expected{ExitCode: 1})

	result = icmd.RunCmd(icmd.Command(GomplateBin,
		"--plugin", "f="+s.tmpDir.Join("fail.bat"),
		"-i", `{{ f "bat failed" 42 }}`,
	), func(c *icmd.Cmd) {})
	result.Assert(c, icmd.Expected{ExitCode: 1, Err: "bat failed"})
	result.Assert(c, icmd.Expected{ExitCode: 1, Err: "error calling f: exit status 42"})
}
