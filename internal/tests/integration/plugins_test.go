//go:build !windows
// +build !windows

package integration

import (
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"
)

func setupPluginsTest(t *testing.T) *fs.Dir {
	tmpDir := fs.NewDir(t, "gomplate-inttests",
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
	t.Cleanup(tmpDir.Remove)

	return tmpDir
}

func TestPlugins(t *testing.T) {
	tmpDir := setupPluginsTest(t)
	o, e, err := cmd(t, "--plugin", "hi="+tmpDir.Join("foo.sh"),
		"-i", `{{ hi "hello world" }}`).run()
	assertSuccess(t, o, e, err, "hello world\n")

	o, e, err = cmd(t, "--plugin", "echo="+tmpDir.Join("bar.sh"),
		"-i", `{{ echo "$HELLO" }}`).
		withEnv("HELLO", "hello world").run()
	assertSuccess(t, o, e, err, "hello world\n")
}

func TestPlugins_Errors(t *testing.T) {
	tmpDir := setupPluginsTest(t)
	_, _, err := cmd(t, "--plugin", "f=false",
		"-i", `{{ f }}`).run()
	assert.ErrorContains(t, err, "exit status 1")

	_, _, err = cmd(t, "--plugin", "f="+tmpDir.Join("fail.sh"),
		"-i", `{{ f "all is lost" 5 }}`).run()
	assert.ErrorContains(t, err, "all is lost")
	assert.ErrorContains(t, err, "error calling f: exit status 5")
}

func TestPlugins_Timeout(t *testing.T) {
	tmpDir := setupPluginsTest(t)
	_, _, err := cmd(t, "--plugin", "sleep="+tmpDir.Join("sleep.sh"),
		"-i", `{{ sleep 10 }}`).run()
	assert.ErrorContains(t, err, "plugin timed out")

	_, _, err = cmd(t, "--plugin", "sleep="+tmpDir.Join("sleep.sh"),
		"-i", `{{ sleep 2 }}`).
		withEnv("GOMPLATE_PLUGIN_TIMEOUT", "500ms").run()
	assert.ErrorContains(t, err, "plugin timed out")
}
