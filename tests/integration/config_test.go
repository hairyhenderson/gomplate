//+build integration

package integration

import (
	"bytes"
	"io/ioutil"
	"os"
	"runtime"

	"gopkg.in/check.v1"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"
	"gotest.tools/v3/icmd"
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
	result := icmd.RunCmd(icmd.Command(GomplateBin), func(cmd *icmd.Cmd) {
		cmd.Dir = s.tmpDir.Path()
	})
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "hello world"})
}

func (s *ConfigSuite) TestReadsStdin(c *check.C) {
	s.writeConfig("inputFiles: [-]")
	result := icmd.RunCmd(icmd.Command(GomplateBin), func(cmd *icmd.Cmd) {
		cmd.Dir = s.tmpDir.Path()
		cmd.Stdin = bytes.NewBufferString("foo bar")
	})
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "foo bar"})
}

func (s *ConfigSuite) TestFlagOverridesConfig(c *check.C) {
	s.writeConfig("inputFiles: [in]")
	result := icmd.RunCmd(icmd.Command(GomplateBin, "-i", "hello from the cli"), func(cmd *icmd.Cmd) {
		cmd.Dir = s.tmpDir.Path()
	})
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "hello from the cli"})
}

func (s *ConfigSuite) TestReadsFromInputFile(c *check.C) {
	s.writeConfig("inputFiles: [in]")
	s.writeFile("in", "blah blah")
	result := icmd.RunCmd(icmd.Command(GomplateBin), func(cmd *icmd.Cmd) {
		cmd.Dir = s.tmpDir.Path()
	})
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "blah blah"})
}

func (s *ConfigSuite) TestDatasource(c *check.C) {
	s.writeConfig(`inputFiles: [in]
datasources:
  data:
    url: in.yaml
`)
	s.writeFile("in", `{{ (ds "data").value }}`)
	s.writeFile("in.yaml", `value: hello world`)
	result := icmd.RunCmd(icmd.Command(GomplateBin), func(cmd *icmd.Cmd) {
		cmd.Dir = s.tmpDir.Path()
	})
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "hello world"})
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
	result := icmd.RunCmd(icmd.Command(GomplateBin), func(cmd *icmd.Cmd) {
		cmd.Dir = s.tmpDir.Path()
	})
	result.Assert(c, icmd.Expected{ExitCode: 0})
	b, err := ioutil.ReadFile(s.tmpDir.Join("outdir", "file"))
	assert.NilError(c, err)
	assert.Equal(c, "hello world", string(b))
}

func (s *ConfigSuite) TestExecPipeOverridesConfigFile(c *check.C) {
	// make sure exec-pipe works, and outFiles is replaced
	s.writeConfig(`in: hello world
outputFiles: ['-']
`)
	result := icmd.RunCmd(icmd.Command(GomplateBin, "-i", "hi", "--exec-pipe", "--", "tr", "[a-z]", "[A-Z]"), func(cmd *icmd.Cmd) {
		cmd.Dir = s.tmpDir.Path()
	})
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "HI"})
}

func (s *ConfigSuite) TestOutFile(c *check.C) {
	s.writeConfig(`in: hello world
outputFiles: [out]
`)
	result := icmd.RunCmd(icmd.Command(GomplateBin), func(cmd *icmd.Cmd) {
		cmd.Dir = s.tmpDir.Path()
	})
	result.Assert(c, icmd.Expected{ExitCode: 0})
	b, err := ioutil.ReadFile(s.tmpDir.Join("out"))
	assert.NilError(c, err)
	assert.Equal(c, "hello world", string(b))
}

func (s *ConfigSuite) TestAlternateConfigFile(c *check.C) {
	s.writeFile("config.yaml", `in: this is from an alternate config
`)
	result := icmd.RunCmd(icmd.Command(GomplateBin, "--config=config.yaml"), func(cmd *icmd.Cmd) {
		cmd.Dir = s.tmpDir.Path()
	})
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "this is from an alternate config"})
}

func (s *ConfigSuite) TestEnvConfigFile(c *check.C) {
	s.writeFile("envconfig.yaml", `in: yet another alternate config
`)
	result := icmd.RunCmd(icmd.Command(GomplateBin), func(cmd *icmd.Cmd) {
		cmd.Dir = s.tmpDir.Path()
		cmd.Env = []string{"GOMPLATE_CONFIG=./envconfig.yaml"}
	})
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "yet another alternate config"})
}

func (s *ConfigSuite) TestConfigOverridesEnvDelim(c *check.C) {
	if runtime.GOOS != "windows" {
		s.writeConfig(`inputFiles: [in]
leftDelim: (╯°□°）╯︵ ┻━┻
datasources:
  data:
    url: in.yaml
`)
		s.writeFile("in", `(╯°□°）╯︵ ┻━┻ (ds "data").value }}`)
		s.writeFile("in.yaml", `value: hello world`)
		result := icmd.RunCmd(icmd.Command(GomplateBin), func(cmd *icmd.Cmd) {
			cmd.Dir = s.tmpDir.Path()
			cmd.Env = []string{"GOMPLATE_LEFT_DELIM", "<<"}
		})
		result.Assert(c, icmd.Expected{ExitCode: 0, Out: "hello world"})
	}
}

func (s *ConfigSuite) TestFlagOverridesAllDelim(c *check.C) {
	if runtime.GOOS != "windows" {
		s.writeConfig(`inputFiles: [in]
leftDelim: (╯°□°）╯︵ ┻━┻
datasources:
  data:
    url: in.yaml
`)
		s.writeFile("in", `{{ (ds "data").value }}`)
		s.writeFile("in.yaml", `value: hello world`)
		result := icmd.RunCmd(icmd.Command(GomplateBin, "--left-delim={{"), func(cmd *icmd.Cmd) {
			cmd.Dir = s.tmpDir.Path()
			cmd.Env = []string{"GOMPLATE_LEFT_DELIM", "<<"}
		})
		result.Assert(c, icmd.Expected{ExitCode: 0, Out: "hello world"})
	}
}

func (s *ConfigSuite) TestConfigOverridesEnvPluginTimeout(c *check.C) {
	if runtime.GOOS != "windows" {
		s.writeConfig(`in: hi there {{ sleep 2 }}
plugins:
  sleep: echo

pluginTimeout: 500ms
`)
		result := icmd.RunCmd(icmd.Command(GomplateBin,
			"--plugin", "sleep="+s.tmpDir.Join("sleep.sh"),
		), func(cmd *icmd.Cmd) {
			cmd.Dir = s.tmpDir.Path()
			cmd.Env = []string{"GOMPLATE_PLUGIN_TIMEOUT=5s"}
		})
		result.Assert(c, icmd.Expected{ExitCode: 1, Err: "plugin timed out"})
	}
}

func (s *ConfigSuite) TestConfigOverridesEnvSuppressEmpty(c *check.C) {
	s.writeConfig(`in: |
  {{- print "\t  \n\n\r\n\t\t     \v\n" -}}

  {{ print "   " -}}
out: ./missing
suppressEmpty: true
`)
	result := icmd.RunCmd(icmd.Command(GomplateBin), func(cmd *icmd.Cmd) {
		cmd.Dir = s.tmpDir.Path()
		// should have no effect, as config overrides
		cmd.Env = []string{"GOMPLATE_SUPPRESS_EMPTY=false"}
	})
	result.Assert(c, icmd.Expected{ExitCode: 0})
	_, err := os.Stat(s.tmpDir.Join("missing"))
	assert.Equal(c, true, os.IsNotExist(err))
}
