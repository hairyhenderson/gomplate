package integration

import (
	"io/fs"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tfs "gotest.tools/v3/fs"
)

func setupConfigTest(t *testing.T) *tfs.Dir {
	t.Helper()

	tmpDir := tfs.NewDir(t, "gomplate-inttests",
		tfs.WithDir("indir"),
		tfs.WithDir("outdir"),
		tfs.WithFile(".gomplate.yaml", "in: hello world\n"),
		tfs.WithFile("sleep.sh", "#!/bin/sh\n\nexec sleep $1\n", tfs.WithMode(0o755)),
	)
	t.Cleanup(tmpDir.Remove)
	return tmpDir
}

func writeFile(t *testing.T, dir *tfs.Dir, f, content string) {
	t.Helper()

	f = dir.Join(f)
	err := os.WriteFile(f, []byte(content), 0o600)
	if err != nil {
		t.Fatal(err)
	}
}

func writeConfig(t *testing.T, dir *tfs.Dir, content string) {
	t.Helper()

	writeFile(t, dir, ".gomplate.yaml", content)
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
	writeFile(t, tmpDir, "in", "blah blah")

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
	writeFile(t, tmpDir, "in", `{{ (ds "data").value }}`)
	writeFile(t, tmpDir, "in.yaml", `value: hello world`)

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
	writeFile(t, tmpDir, "indir/file", `{{ (ds "data").value }}`)
	writeFile(t, tmpDir, "in.yaml", `value: hello world`)

	o, e, err := cmd(t).withDir(tmpDir.Path()).run()
	assertSuccess(t, o, e, err, "")

	b, err := os.ReadFile(tmpDir.Join("outdir", "file"))
	require.NoError(t, err)
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

	b, err := os.ReadFile(tmpDir.Join("out"))
	require.NoError(t, err)
	assert.Equal(t, "hello world", string(b))
}

func TestConfig_AlternateConfigFile(t *testing.T) {
	tmpDir := setupConfigTest(t)
	writeFile(t, tmpDir, "config.yaml", `in: this is from an alternate config
`)

	o, e, err := cmd(t, "--config=config.yaml").withDir(tmpDir.Path()).run()
	assertSuccess(t, o, e, err, "this is from an alternate config")
}

func TestConfig_EnvConfigFile(t *testing.T) {
	tmpDir := setupConfigTest(t)

	writeFile(t, tmpDir, "envconfig.yaml", `in: yet another alternate config
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
	writeFile(t, tmpDir, "in", `(╯°□°）╯︵ ┻━┻ (ds "data").value }}`)
	writeFile(t, tmpDir, "in.yaml", `value: hello world`)

	o, e, err := cmd(t).withDir(tmpDir.Path()).
		withEnv("GOMPLATE_LEFT_DELIM", "<<").run()
	require.NoError(t, err)
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
	writeFile(t, tmpDir, "in", `{{ (ds "data").value }}`)
	writeFile(t, tmpDir, "in.yaml", `value: hello world`)

	o, e, err := cmd(t, "--left-delim={{").
		withDir(tmpDir.Path()).
		withEnv("GOMPLATE_LEFT_DELIM", "<<").run()
	require.NoError(t, err)
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
	require.NoError(t, err)

	_, err = os.Stat(tmpDir.Join("missing"))
	require.ErrorIs(t, err, fs.ErrNotExist)
}

func TestConfig_ConfigParseErrorSpecifiesFilename(t *testing.T) {
	tmpDir := setupConfigTest(t)

	writeConfig(t, tmpDir, `templates:
  dir: /foo/bar
`)
	_, _, err := cmd(t).withDir(tmpDir.Path()).run()
	assert.ErrorContains(t, err, `parsing config file ".gomplate.yaml": YAML decoding failed`)
}

func TestConfig_ConfigTemplatesSupportsMap(t *testing.T) {
	tmpDir := setupConfigTest(t)

	writeConfig(t, tmpDir, `in: '{{ template "t1" (dict "testValue" "12345") }}'
templates:
  t1:
    url: t1.tmpl
`)
	writeFile(t, tmpDir, "t1.tmpl", `{{ .testValue }}`)

	o, e, err := cmd(t).withDir(tmpDir.Path()).run()
	assertSuccess(t, o, e, err, "12345")
}

func TestConfig_ConfigTemplatesSupportsArray(t *testing.T) {
	tmpDir := setupConfigTest(t)

	// TODO: remove this test once the array format is no longer supported
	writeConfig(t, tmpDir, `in: '{{ template "t1" (dict "testValue" "12345") }}'
templates:
  - t1=t1.tmpl
`)
	writeFile(t, tmpDir, "t1.tmpl", `{{ .testValue }}`)

	o, e, err := cmd(t).withDir(tmpDir.Path()).run()
	assert.Contains(t, e, "Deprecated: config: the YAML array form for 'templates' is deprecated")
	assert.Equal(t, "12345", o)
	require.NoError(t, err)
}

func TestConfig_MissingKeyDefault(t *testing.T) {
	tmpDir := setupConfigTest(t)
	writeConfig(t, tmpDir, `inputFiles: [in]
missingKey: default
`)
	writeFile(t, tmpDir, "in", `{{ .name }}`)

	o, e, err := cmd(t).withDir(tmpDir.Path()).run()
	assertSuccess(t, o, e, err, `<no value>`)
}

func TestConfig_MissingKeyNotDefined(t *testing.T) {
	tmpDir := setupConfigTest(t)
	writeConfig(t, tmpDir, `inputFiles: [in]`)
	writeFile(t, tmpDir, "in", `{{ .name }}`)

	o, e, err := cmd(t).withDir(tmpDir.Path()).run()
	assertFailed(t, o, e, err, `map has no entry for key \"name\"`)
}
