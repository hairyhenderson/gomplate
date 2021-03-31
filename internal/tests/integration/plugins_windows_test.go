//+build windows

package integration

import (
	"strings"
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"
)

func setupPluginsTest(t *testing.T) *fs.Dir {
	tmpDir := fs.NewDir(t, "gomplate-inttests",
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
	t.Cleanup(tmpDir.Remove)

	return tmpDir
}

func TestPlugins(t *testing.T) {
	tmpDir := setupPluginsTest(t)
	o, e, err := cmd(t,
		"--plugin", "foo="+tmpDir.Join("foo.bat"),
		"-i", `{{ foo "hello world" }}`,
	).run()
	assertSuccess(t, strings.TrimSpace(o), e, err, `"hello world"`)
}

func TestPlugins_Errors(t *testing.T) {
	tmpDir := setupPluginsTest(t)
	_, _, err := cmd(t, "--plugin", "f=false",
		"-i", `{{ f }}`).run()
	assert.ErrorContains(t, err, "exit status 1")

	_, _, err = cmd(t, "--plugin", "f="+tmpDir.Join("fail.bat"),
		"-i", `{{ f "bat failed" 42 }}`).run()
	assert.ErrorContains(t, err, "bat failed")
	assert.ErrorContains(t, err, "error calling f: exit status 42")
}
