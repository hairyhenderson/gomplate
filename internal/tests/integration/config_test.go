package integration

import (
	"io/ioutil"
	"os"
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"
)

func setupConfigTest(t *testing.T) *fs.Dir {
	tmpDir := fs.NewDir(t, "gomplate-inttests",
		fs.WithDir("indir"),
		fs.WithDir("outdir"),
		fs.WithFile(".gomplate.yaml", "in: hello world\n"),
		fs.WithFile("sleep.sh", "#!/bin/sh\n\nexec sleep $1\n", fs.WithMode(0755)),
	)
	t.Cleanup(tmpDir.Remove)
	return tmpDir
}

func writeFile(dir *fs.Dir, f, content string) {
	f = dir.Join(f)
	err := ioutil.WriteFile(f, []byte(content), 0600)
	if err != nil {
		panic(err)
	}
}

func writeConfig(t *testing.T, dir *fs.Dir, content string) {
	t.Helper()

	writeFile(dir, ".gomplate.yaml", content)
	t.Logf("writing config: %s", content)
}

func TestConfig_ReadsFromSimpleConfigFile(t *testing.T) {
	tmpDir := setupConfigTest(t)

	o, e, err := cmd(t).withDir(tmpDir.Path()).run()
	assertSuccess(t, o, e, err, "hello world")
}

func TestConfig_ReadsStdin(t *testing.T) {
	tmpDir := setupConfigTest(t)
	writeConfig(t, tmpDir, "inputFiles: [-]")

	o, e, err := cmd(t).withDir(tmpDir.Path()).withStdin("foo bar").run()
	assertSuccess(t, o, e, err, "foo bar")
}

func TestConfig_FlagOverridesConfig(t *testing.T) {
	tmpDir := setupConfigTest(t)
	writeConfig(t, tmpDir, "inputFiles: [in]")

	o, e, err := cmd(t, "-i", "hello from the cli").
		withDir(tmpDir.Path()).run()
	assertSuccess(t, o, e, err, "hello from the cli")
}

func TestConfig_ReadsFromInputFile(t *testing.T) {
	tmpDir := setupConfigTest(t)
	writeConfig(t, tmpDir, "inputFiles: [in]")
	writeFile(tmpDir, "in", "blah blah")

	o, e, err := cmd(t).withDir(tmpDir.Path()).run()
	assertSuccess(t, o, e, err, "blah blah")
}

func TestConfig_Datasource(t *testing.T) {
	tmpDir := setupConfigTest(t)
	writeConfig(t, tmpDir, `inputFiles: [in]
datasources:
  data:
    url: in.yaml
`)
	writeFile(tmpDir, "in", `{{ (ds "data").value }}`)
	writeFile(tmpDir, "in.yaml", `value: hello world`)

	o, e, err := cmd(t).withDir(tmpDir.Path()).run()
	assertSuccess(t, o, e, err, "hello world")
}

func TestConfig_OutputDir(t *testing.T) {
	tmpDir := setupConfigTest(t)

	writeConfig(t, tmpDir, `inputDir: indir/
outputDir: outdir/
datasources:
  data:
    url: in.yaml
`)
	writeFile(tmpDir, "indir/file", `{{ (ds "data").value }}`)
	writeFile(tmpDir, "in.yaml", `value: hello world`)

	o, e, err := cmd(t).withDir(tmpDir.Path()).run()
	assertSuccess(t, o, e, err, "")

	b, err := ioutil.ReadFile(tmpDir.Join("outdir", "file"))
	assert.NilError(t, err)
	assert.Equal(t, "hello world", string(b))
}

func TestConfig_ExecPipeOverridesConfigFile(t *testing.T) {
	tmpDir := setupConfigTest(t)

	// make sure exec-pipe works, and outFiles is replaced
	writeConfig(t, tmpDir, `in: hello world
outputFiles: ['-']
`)
	o, e, err := cmd(t, "-i", "hi", "--exec-pipe", "--", "tr", "[a-z]", "[A-Z]").
		withDir(tmpDir.Path()).run()
	assertSuccess(t, o, e, err, "HI")
}

func TestConfig_OutFile(t *testing.T) {
	tmpDir := setupConfigTest(t)

	writeConfig(t, tmpDir, `in: hello world
outputFiles: [out]
`)
	o, e, err := cmd(t).withDir(tmpDir.Path()).run()
	assertSuccess(t, o, e, err, "")

	b, err := ioutil.ReadFile(tmpDir.Join("out"))
	assert.NilError(t, err)
	assert.Equal(t, "hello world", string(b))
}

func TestConfig_AlternateConfigFile(t *testing.T) {
	tmpDir := setupConfigTest(t)
	writeFile(tmpDir, "config.yaml", `in: this is from an alternate config
`)

	o, e, err := cmd(t, "--config=config.yaml").withDir(tmpDir.Path()).run()
	assertSuccess(t, o, e, err, "this is from an alternate config")
}

func TestConfig_EnvConfigFile(t *testing.T) {
	tmpDir := setupConfigTest(t)

	writeFile(tmpDir, "envconfig.yaml", `in: yet another alternate config
`)

	o, e, err := cmd(t).withDir(tmpDir.Path()).
		withEnv("GOMPLATE_CONFIG", "./envconfig.yaml").run()
	assertSuccess(t, o, e, err, "yet another alternate config")
}

func TestConfig_ConfigOverridesEnvDelim(t *testing.T) {
	if isWindows {
		t.Skip()
	}

	tmpDir := setupConfigTest(t)

	writeConfig(t, tmpDir, `inputFiles: [in]
leftDelim: (╯°□°）╯︵ ┻━┻
datasources:
  data:
    url: in.yaml
`)
	writeFile(tmpDir, "in", `(╯°□°）╯︵ ┻━┻ (ds "data").value }}`)
	writeFile(tmpDir, "in.yaml", `value: hello world`)

	o, e, err := cmd(t).withDir(tmpDir.Path()).
		withEnv("GOMPLATE_LEFT_DELIM", "<<").run()
	assert.NilError(t, err)
	assert.Equal(t, "", e)
	assert.Equal(t, "hello world", o)
}

func TestConfig_FlagOverridesAllDelim(t *testing.T) {
	if isWindows {
		t.Skip()
	}

	tmpDir := setupConfigTest(t)

	writeConfig(t, tmpDir, `inputFiles: [in]
leftDelim: (╯°□°）╯︵ ┻━┻
datasources:
  data:
    url: in.yaml
`)
	writeFile(tmpDir, "in", `{{ (ds "data").value }}`)
	writeFile(tmpDir, "in.yaml", `value: hello world`)

	o, e, err := cmd(t, "--left-delim={{").
		withDir(tmpDir.Path()).
		withEnv("GOMPLATE_LEFT_DELIM", "<<").run()
	assert.NilError(t, err)
	assert.Equal(t, "", e)
	assert.Equal(t, "hello world", o)
}

func TestConfig_ConfigOverridesEnvPluginTimeout(t *testing.T) {
	if isWindows {
		t.Skip()
	}

	tmpDir := setupConfigTest(t)

	writeConfig(t, tmpDir, `in: hi there {{ sleep 2 }}
plugins:
  sleep: echo

pluginTimeout: 500ms
`)

	_, _, err := cmd(t, "--plugin", "sleep="+tmpDir.Join("sleep.sh")).
		withDir(tmpDir.Path()).
		withEnv("GOMPLATE_PLUGIN_TIMEOUT", "5s").run()
	assert.ErrorContains(t, err, "plugin timed out")
}

func TestConfig_ConfigOverridesEnvSuppressEmpty(t *testing.T) {
	tmpDir := setupConfigTest(t)

	writeConfig(t, tmpDir, `in: |
  {{- print "\t  \n\n\r\n\t\t     \v\n" -}}

  {{ print "   " -}}
out: ./missing
suppressEmpty: true
`)

	_, _, err := cmd(t).withDir(tmpDir.Path()).
		withEnv("GOMPLATE_SUPPRESS_EMPTY", "false").run()
	assert.NilError(t, err)

	_, err = os.Stat(tmpDir.Join("missing"))
	assert.Equal(t, true, os.IsNotExist(err))
}
