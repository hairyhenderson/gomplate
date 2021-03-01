//+build windows

package integration

import (
	"strings"

	. "gopkg.in/check.v1"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"
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
	o, e, err := cmdTest(c,
		"--plugin", "foo="+s.tmpDir.Join("foo.bat"),
		"-i", `{{ foo "hello world" }}`,
	)
	assertSuccess(c, strings.TrimSpace(o), e, err, `"hello world"`)
}

func (s *PluginsSuite) TestPluginErrors(c *C) {
	_, _, err := cmdTest(c, "--plugin", "f=false",
		"-i", `{{ f }}`)
	assert.ErrorContains(c, err, "exit status 1")

	_, _, err = cmdTest(c, "--plugin", "f="+s.tmpDir.Join("fail.bat"),
		"-i", `{{ f "bat failed" 42 }}`)
	assert.ErrorContains(c, err, "bat failed")
	assert.ErrorContains(c, err, "error calling f: exit status 42")
}
