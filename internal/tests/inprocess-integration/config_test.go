package integration

import (
	"io/ioutil"
	"os"

	"gopkg.in/check.v1"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"
)

type ConfigSuite struct {
	tmpDir *fs.Dir
}

var _ = check.Suite(&ConfigSuite{})

func (s *ConfigSuite) SetUpTest(c *check.C) {
	s.tmpDir = fs.NewDir(c, "gomplate-inttests",
		fs.WithDir("indir"),
		fs.WithDir("outdir"),
		fs.WithFile(".gomplate.yaml", "in: hello world\n"),
		fs.WithFile("sleep.sh", "#!/bin/sh\n\nexec sleep $1\n", fs.WithMode(0755)),
	)
}

func (s *ConfigSuite) writeFile(f, content string) {
	f = s.tmpDir.Join(f)
	err := ioutil.WriteFile(f, []byte(content), 0600)
	if err != nil {
		panic(err)
	}
}

func (s *ConfigSuite) writeConfig(content string) {
	s.writeFile(".gomplate.yaml", content)
}

func (s *ConfigSuite) TearDownTest(c *check.C) {
	s.tmpDir.Remove()
}

func (s *ConfigSuite) TestReadsFromSimpleConfigFile(c *check.C) {
	o, e, err := cmdWithDir(c, s.tmpDir.Path())
	assertSuccess(c, o, e, err, "hello world")
}

func (s *ConfigSuite) TestReadsStdin(c *check.C) {
	s.writeConfig("inputFiles: [-]")

	origWd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	defer func() { os.Chdir(origWd) }()

	err = os.Chdir(s.tmpDir.Path())
	if err != nil {
		panic(err)
	}

	o, e, err := cmdWithStdin(c, nil, "foo bar")
	assertSuccess(c, o, e, err, "foo bar")
}

func (s *ConfigSuite) TestFlagOverridesConfig(c *check.C) {
	s.writeConfig("inputFiles: [in]")

	o, e, err := cmdWithDir(c, s.tmpDir.Path(), "-i", "hello from the cli")
	assertSuccess(c, o, e, err, "hello from the cli")
}

func (s *ConfigSuite) TestReadsFromInputFile(c *check.C) {
	s.writeConfig("inputFiles: [in]")
	s.writeFile("in", "blah blah")

	o, e, err := cmdWithDir(c, s.tmpDir.Path())
	assertSuccess(c, o, e, err, "blah blah")
}

func (s *ConfigSuite) TestDatasource(c *check.C) {
	s.writeConfig(`inputFiles: [in]
datasources:
  data:
    url: in.yaml
`)
	s.writeFile("in", `{{ (ds "data").value }}`)
	s.writeFile("in.yaml", `value: hello world`)

	o, e, err := cmdWithDir(c, s.tmpDir.Path())
	assertSuccess(c, o, e, err, "hello world")
}

func (s *ConfigSuite) TestOutputDir(c *check.C) {
	s.writeConfig(`inputDir: indir/
outputDir: outdir/
datasources:
  data:
    url: in.yaml
`)
	s.writeFile("indir/file", `{{ (ds "data").value }}`)
	s.writeFile("in.yaml", `value: hello world`)

	o, e, err := cmdWithDir(c, s.tmpDir.Path())
	assertSuccess(c, o, e, err, "")

	b, err := ioutil.ReadFile(s.tmpDir.Join("outdir", "file"))
	assert.NilError(c, err)
	assert.Equal(c, "hello world", string(b))
}

func (s *ConfigSuite) TestExecPipeOverridesConfigFile(c *check.C) {
	// make sure exec-pipe works, and outFiles is replaced
	s.writeConfig(`in: hello world
outputFiles: ['-']
`)
	o, e, err := cmdWithDir(c, s.tmpDir.Path(), "-i", "hi", "--exec-pipe", "--", "tr", "[a-z]", "[A-Z]")
	assertSuccess(c, o, e, err, "HI")
}

func (s *ConfigSuite) TestOutFile(c *check.C) {
	s.writeConfig(`in: hello world
outputFiles: [out]
`)
	o, e, err := cmdWithDir(c, s.tmpDir.Path())
	assertSuccess(c, o, e, err, "")

	b, err := ioutil.ReadFile(s.tmpDir.Join("out"))
	assert.NilError(c, err)
	assert.Equal(c, "hello world", string(b))
}

func (s *ConfigSuite) TestAlternateConfigFile(c *check.C) {
	s.writeFile("config.yaml", `in: this is from an alternate config
`)

	o, e, err := cmdWithDir(c, s.tmpDir.Path(), "--config=config.yaml")
	assertSuccess(c, o, e, err, "this is from an alternate config")
}

func (s *ConfigSuite) TestEnvConfigFile(c *check.C) {
	s.writeFile("envconfig.yaml", `in: yet another alternate config
`)

	os.Setenv("GOMPLATE_CONFIG", "./envconfig.yaml")
	defer os.Unsetenv("GOMPLATE_CONFIG")

	o, e, err := cmdWithDir(c, s.tmpDir.Path())
	assertSuccess(c, o, e, err, "yet another alternate config")
}

func (s *ConfigSuite) TestConfigOverridesEnvDelim(c *check.C) {
	if !isWindows {
		s.writeConfig(`inputFiles: [in]
leftDelim: (╯°□°）╯︵ ┻━┻
datasources:
  data:
    url: in.yaml
`)
		s.writeFile("in", `(╯°□°）╯︵ ┻━┻ (ds "data").value }}`)
		s.writeFile("in.yaml", `value: hello world`)

		os.Setenv("GOMPLATE_LEFT_DELIM", "<<")
		defer os.Unsetenv("GOMPLATE_LEFT_DELIM")

		o, e, err := cmdWithDir(c, s.tmpDir.Path())
		assert.NilError(c, err)
		assert.Equal(c, "", e)
		assert.Equal(c, "hello world", o)
	}
}

func (s *ConfigSuite) TestFlagOverridesAllDelim(c *check.C) {
	if !isWindows {
		s.writeConfig(`inputFiles: [in]
leftDelim: (╯°□°）╯︵ ┻━┻
datasources:
  data:
    url: in.yaml
`)
		s.writeFile("in", `{{ (ds "data").value }}`)
		s.writeFile("in.yaml", `value: hello world`)

		o, e, err := cmdWithDir(c, s.tmpDir.Path(), "--left-delim={{")
		assert.NilError(c, err)
		assert.Equal(c, "", e)
		assert.Equal(c, "hello world", o)
	}
}

func (s *ConfigSuite) TestConfigOverridesEnvPluginTimeout(c *check.C) {
	if !isWindows {
		s.writeConfig(`in: hi there {{ sleep 2 }}
plugins:
  sleep: echo

pluginTimeout: 500ms
`)
		os.Setenv("GOMPLATE_PLUGIN_TIMEOUT", "5s")
		defer os.Unsetenv("GOMPLATE_PLUGIN_TIMEOUT")

		_, _, err := cmdWithDir(c, s.tmpDir.Path(), "--plugin", "sleep="+s.tmpDir.Join("sleep.sh"))
		assert.ErrorContains(c, err, "plugin timed out")
	}
}

func (s *ConfigSuite) TestConfigOverridesEnvSuppressEmpty(c *check.C) {
	s.writeConfig(`in: |
  {{- print "\t  \n\n\r\n\t\t     \v\n" -}}

  {{ print "   " -}}
out: ./missing
suppressEmpty: true
`)

	os.Setenv("GOMPLATE_SUPPRESS_EMPTY", "false")
	defer os.Unsetenv("GOMPLATE_SUPPRESS_EMPTY")

	_, _, err := cmdWithDir(c, s.tmpDir.Path())
	assert.NilError(c, err)

	_, err = os.Stat(s.tmpDir.Join("missing"))
	assert.Equal(c, true, os.IsNotExist(err))
}
