//+build !windows

package integration

import (
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
	o, e, err := cmdTest(c, "--plugin", "hi="+s.tmpDir.Join("foo.sh"),
		"-i", `{{ hi "hello world" }}`)
	assertSuccess(c, o, e, err, "hello world\n")

	cmdWithEnv(c, []string{
		"--plugin", "echo=" + s.tmpDir.Join("bar.sh"),
		"-i", `{{ echo "$HELLO" }}`,
	}, map[string]string{"HELLO": "hello world"})
	assertSuccess(c, o, e, err, "hello world\n")
}

func (s *PluginsSuite) TestPluginErrors(c *C) {
	_, _, err := cmdTest(c, "--plugin", "f=false",
		"-i", `{{ f }}`)
	assert.ErrorContains(c, err, "exit status 1")

	_, _, err = cmdTest(c, "--plugin", "f="+s.tmpDir.Join("fail.sh"),
		"-i", `{{ f "all is lost" 5 }}`)
	assert.ErrorContains(c, err, "all is lost")
	assert.ErrorContains(c, err, "error calling f: exit status 5")
}

func (s *PluginsSuite) TestPluginTimeout(c *C) {
	_, _, err := cmdTest(c, "--plugin", "sleep="+s.tmpDir.Join("sleep.sh"),
		"-i", `{{ sleep 10 }}`,
	)
	assert.ErrorContains(c, err, "plugin timed out")

	_, _, err = cmdWithEnv(c, []string{
		"--plugin", "sleep=" + s.tmpDir.Join("sleep.sh"),
		"-i", `{{ sleep 2 }}`,
	}, map[string]string{"GOMPLATE_PLUGIN_TIMEOUT": "500ms"})
	assert.ErrorContains(c, err, "plugin timed out")
}
