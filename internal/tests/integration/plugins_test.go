//+build integration
//+build !windows

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
		fs.WithFile("foo.sh", "#!/bin/sh\n\necho $*\n", fs.WithMode(0755)),
		fs.WithFile("foo.ps1", "echo $args\r\n", fs.WithMode(0755)),
		fs.WithFile("bar.sh", "#!/bin/sh\n\neval \"echo $*\"\n", fs.WithMode(0755)),
		fs.WithFile("fail.sh", "#!/bin/sh\n\n>&2 echo $1\nexit $2\n", fs.WithMode(0755)),
		fs.WithFile("fail.ps1", `param (
	[Parameter(Position=0)]
	[string]$msg,

	[Parameter(Position=1)]
	[int]$code
)
write-error $msg
exit $code
`, fs.WithMode(0755)),
		fs.WithFile("sleep.sh", "#!/bin/sh\n\nexec sleep $1\n", fs.WithMode(0755)),
	)
}

func (s *PluginsSuite) TearDownSuite(c *C) {
	s.tmpDir.Remove()
}

func (s *PluginsSuite) TestPlugins(c *C) {
	result := icmd.RunCommand(GomplateBin,
		"--plugin", "hi="+s.tmpDir.Join("foo.sh"),
		"-i", `{{ hi "hello world" }}`,
	)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "hello world"})

	result = icmd.RunCmd(icmd.Command(GomplateBin,
		"--plugin", "echo="+s.tmpDir.Join("bar.sh"),
		"-i", `{{ echo "$HELLO" }}`,
	), func(c *icmd.Cmd) {
		c.Env = []string{
			"HELLO=hello world",
		}
	})
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "hello world"})
}

func (s *PluginsSuite) TestPluginErrors(c *C) {
	result := icmd.RunCmd(icmd.Command(GomplateBin,
		"--plugin", "f=false",
		"-i", `{{ f }}`,
	), func(c *icmd.Cmd) {})
	result.Assert(c, icmd.Expected{ExitCode: 1})

	result = icmd.RunCmd(icmd.Command(GomplateBin,
		"--plugin", "f="+s.tmpDir.Join("fail.sh"),
		"-i", `{{ f "all is lost" 5 }}`,
	), func(c *icmd.Cmd) {})
	result.Assert(c, icmd.Expected{ExitCode: 1, Err: "all is lost"})
	result.Assert(c, icmd.Expected{ExitCode: 1, Err: "error calling f: exit status 5"})
}

func (s *PluginsSuite) TestPluginTimeout(c *C) {
	result := icmd.RunCmd(icmd.Command(GomplateBin,
		"--plugin", "sleep="+s.tmpDir.Join("sleep.sh"),
		"-i", `{{ sleep 10 }}`,
	), func(c *icmd.Cmd) {})
	result.Assert(c, icmd.Expected{ExitCode: 1, Err: "plugin timed out"})

	result = icmd.RunCmd(icmd.Command(GomplateBin,
		"--plugin", "sleep="+s.tmpDir.Join("sleep.sh"),
		"-i", `{{ sleep 2 }}`,
	), func(c *icmd.Cmd) {
		c.Env = []string{"GOMPLATE_PLUGIN_TIMEOUT=500ms"}
	})
	result.Assert(c, icmd.Expected{ExitCode: 1, Err: "plugin timed out"})
}
