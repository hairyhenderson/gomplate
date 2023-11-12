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
		fs.WithFile("foo.sh", "#!/bin/sh\n\necho $*\n", fs.WithMode(0o755)),
		fs.WithFile("foo.ps1", "echo $args\r\n", fs.WithMode(0o755)),
		fs.WithFile("bar.sh", "#!/bin/sh\n\neval \"echo $*\"\n", fs.WithMode(0o755)),
		fs.WithFile("fail.sh", "#!/bin/sh\n\n>&2 echo $1\nexit $2\n", fs.WithMode(0o755)),
		fs.WithFile("fail.ps1", `param (
	[Parameter(Position=0)]
	[string]$msg,

	[Parameter(Position=1)]
	[int]$code
)
write-error $msg
exit $code
`, fs.WithMode(0o755)),
		fs.WithFile("sleep.sh", "#!/bin/sh\n\nexec sleep $1\n", fs.WithMode(0o755)),
		fs.WithFile("replace.sh", `#!/bin/sh
if [ "$#" -eq 2 ]; then
	exec tr $1 $2
elif [ "$#" -eq 3 ]; then
	printf "=%s" $3 | tr $1 $2
fi
`, fs.WithMode(0o755)),
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
	if testing.Short() {
		t.Skip()
	}

	tmpDir := setupPluginsTest(t)
	_, _, err := cmd(t, "--plugin", "sleep="+tmpDir.Join("sleep.sh"),
		"-i", `{{ sleep 10 }}`).run()
	assert.ErrorContains(t, err, "plugin timed out")

	_, _, err = cmd(t, "--plugin", "sleep="+tmpDir.Join("sleep.sh"),
		"-i", `{{ sleep 2 }}`).
		withEnv("GOMPLATE_PLUGIN_TIMEOUT", "500ms").run()
	assert.ErrorContains(t, err, "plugin timed out")
}

func TestPlugins_PipeMode(t *testing.T) {
	tmpDir := setupPluginsTest(t)

	writeConfig(t, tmpDir, `in: '{{ "hi there" | replace "h" "H" }}'
plugins:
  replace:
    cmd: `+tmpDir.Join("replace.sh")+`
    pipe: true
`)

	o, e, err := cmd(t).withDir(tmpDir.Path()).run()
	assert.NilError(t, err)
	assert.Equal(t, "", e)
	assert.Equal(t, "Hi tHere", o)

	writeConfig(t, tmpDir, `in: '{{ "hi there" | replace "e" "Z" }}'
plugins:
  replace:
    cmd: `+tmpDir.Join("replace.sh")+`
`)

	o, e, err = cmd(t).withDir(tmpDir.Path()).run()
	assert.NilError(t, err)
	assert.Equal(t, "", e)
	assert.Equal(t, "=hi=thZrZ", o)
}
