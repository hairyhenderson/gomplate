//+build integration

package integration

import (
	"bytes"
	"io/ioutil"
	"os"

	"github.com/hairyhenderson/gomplate/v3/internal/config"

	. "gopkg.in/check.v1"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
	"gotest.tools/v3/fs"
	"gotest.tools/v3/icmd"
)

type BasicSuite struct {
	tmpDir *fs.Dir
}

var _ = Suite(&BasicSuite{})

func (s *BasicSuite) SetUpTest(c *C) {
	s.tmpDir = fs.NewDir(c, "gomplate-inttests",
		fs.WithFile("one", "hi\n", fs.WithMode(0640)),
		fs.WithFile("two", "hello\n"),
		fs.WithFile("broken", "", fs.WithMode(0000)))
}

func (s *BasicSuite) TearDownTest(c *C) {
	s.tmpDir.Remove()
}

func (s *BasicSuite) TestReportsVersion(c *C) {
	result := icmd.RunCommand(GomplateBin, "-v")
	result.Assert(c, icmd.Success)
	assert.Assert(c, cmp.Contains(result.Combined(), "gomplate version "))
}

func (s *BasicSuite) TestTakesStdinByDefault(c *C) {
	result := icmd.RunCmd(icmd.Command(GomplateBin), func(cmd *icmd.Cmd) {
		cmd.Stdin = bytes.NewBufferString("hello world")
	})
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "hello world"})
	assert.Equal(c, "hello world\n", result.Combined())
}

func (s *BasicSuite) TestTakesStdinWithFileFlag(c *C) {
	result := icmd.RunCmd(icmd.Command(GomplateBin, "--file", "-"), func(cmd *icmd.Cmd) {
		cmd.Stdin = bytes.NewBufferString("hello world")
	})
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "hello world"})
	assert.Equal(c, "hello world\n", result.Combined())
}
func (s *BasicSuite) TestWritesToStdoutWithOutFlag(c *C) {
	result := icmd.RunCmd(icmd.Command(GomplateBin, "--out", "-"), func(cmd *icmd.Cmd) {
		cmd.Stdin = bytes.NewBufferString("hello world")
	})
	assert.Equal(c, 0, result.ExitCode)
	assert.Equal(c, "hello world", result.Stdout())
	assert.Equal(c, "\n", result.Stderr())
}

func (s *BasicSuite) TestIgnoresStdinWithInFlag(c *C) {
	result := icmd.RunCmd(icmd.Command(GomplateBin, "--in", "hi"), func(cmd *icmd.Cmd) {
		cmd.Stdin = bytes.NewBufferString("hello world")
	})
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "hi"})
	assert.Equal(c, "hi\n", result.Combined())
}

func (s *BasicSuite) TestErrorsWithInputOutputImbalance(c *C) {
	result := icmd.RunCmd(icmd.Command(GomplateBin,
		"-f", s.tmpDir.Join("one"),
		"-f", s.tmpDir.Join("two"),
		"-o", s.tmpDir.Join("out")), func(cmd *icmd.Cmd) {
		cmd.Stdin = bytes.NewBufferString("hello world")
	})
	result.Assert(c, icmd.Expected{
		ExitCode: 1,
		Err:      "must provide same number of 'outputFiles' (1) as 'in' or 'inputFiles' (2) options",
	})
}

func (s *BasicSuite) TestRoutesInputsToProperOutputs(c *C) {
	oneOut := s.tmpDir.Join("one.out")
	twoOut := s.tmpDir.Join("two.out")
	result := icmd.RunCmd(icmd.Command(GomplateBin,
		"-f", s.tmpDir.Join("one"),
		"-f", s.tmpDir.Join("two"),
		"-o", oneOut,
		"-o", twoOut), func(cmd *icmd.Cmd) {
		cmd.Stdin = bytes.NewBufferString("hello world")
	})
	result.Assert(c, icmd.Success)
	assert.Equal(c, "", result.Combined())

	testdata := []struct {
		path    string
		mode    os.FileMode
		content string
	}{
		{oneOut, 0640, "hi\n"},
		{twoOut, 0644, "hello\n"},
	}
	for _, v := range testdata {
		info, err := os.Stat(v.path)
		assert.NilError(c, err)
		m := config.NormalizeFileMode(v.mode)
		assert.Equal(c, m, info.Mode(), v.path)
		content, err := ioutil.ReadFile(v.path)
		assert.NilError(c, err)
		assert.Equal(c, v.content, string(content))
	}
}

func (s *BasicSuite) TestFlagRules(c *C) {
	result := icmd.RunCommand(GomplateBin, "-f", "-", "-i", "HELLO WORLD")
	result.Assert(c, icmd.Expected{
		ExitCode: 1,
		Err:      "only one of these options is supported at a time: 'in', 'inputFiles'",
	})

	result = icmd.RunCommand(GomplateBin, "--output-dir", ".")
	result.Assert(c, icmd.Expected{
		ExitCode: 1,
		Err:      "these options must be set together: 'outputDir', 'inputDir'",
	})

	result = icmd.RunCommand(GomplateBin, "--input-dir", ".", "--in", "param")
	result.Assert(c, icmd.Expected{
		ExitCode: 1,
		Err:      "only one of these options is supported at a time: 'in', 'inputDir'",
	})

	result = icmd.RunCommand(GomplateBin, "--input-dir", ".", "--file", "input.txt")
	result.Assert(c, icmd.Expected{
		ExitCode: 1,
		Err:      "only one of these options is supported at a time: 'inputFiles', 'inputDir'",
	})

	result = icmd.RunCommand(GomplateBin, "--output-dir", ".", "--out", "param")
	result.Assert(c, icmd.Expected{
		ExitCode: 1,
		Err:      "only one of these options is supported at a time: 'outputFiles', 'outputDir'",
	})

	result = icmd.RunCommand(GomplateBin, "--output-map", ".", "--out", "param")
	result.Assert(c, icmd.Expected{
		ExitCode: 1,
		Err:      "only one of these options is supported at a time: 'outputFiles', 'outputMap'",
	})
}

func (s *BasicSuite) TestDelimsChangedThroughOpts(c *C) {
	result := icmd.RunCommand(GomplateBin,
		"--left-delim", "((",
		"--right-delim", "))",
		"-i", `foo((print "hi"))`)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "foohi"})
}

func (s *BasicSuite) TestDelimsChangedThroughEnvVars(c *C) {
	result := icmd.RunCmd(icmd.Command(GomplateBin, "-i", `foo<<print "hi">>`),
		func(cmd *icmd.Cmd) {
			cmd.Env = []string{
				"GOMPLATE_LEFT_DELIM=<<",
				"GOMPLATE_RIGHT_DELIM=>>",
			}
		})
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "foohi"})
}

func (s *BasicSuite) TestUnknownArgErrors(c *C) {
	result := icmd.RunCommand(GomplateBin, "-in", "flibbit")
	result.Assert(c, icmd.Expected{ExitCode: 1, Err: `unknown command "flibbit" for "gomplate"`})
}

func (s *BasicSuite) TestExecCommand(c *C) {
	out := s.tmpDir.Join("out")
	result := icmd.RunCmd(icmd.Command(GomplateBin,
		"-i", `{{print "hello world"}}`,
		"-o", out,
		"--", "cat", out))
	result.Assert(c, icmd.Expected{
		ExitCode: 0,
		Out:      "hello world",
	})
	assert.Equal(c, "hello world", result.Combined())
}

func (s *BasicSuite) TestPostRunExecPipe(c *C) {
	result := icmd.RunCmd(icmd.Command(GomplateBin,
		"-i", `{{print "hello world"}}`,
		"--exec-pipe",
		"--", "tr", "a-z", "A-Z"))
	result.Assert(c, icmd.Expected{
		ExitCode: 0,
		Out:      "HELLO WORLD",
	})
	assert.Equal(c, "HELLO WORLD", result.Combined())
}

func (s *BasicSuite) TestEmptyOutputSuppression(c *C) {
	out := s.tmpDir.Join("out")
	result := icmd.RunCmd(icmd.Command(GomplateBin,
		"-i",
		`{{print "\t  \n\n\r\n\t\t     \v\n"}}`,
		"-o", out),
		func(cmd *icmd.Cmd) {
			cmd.Env = []string{
				"GOMPLATE_SUPPRESS_EMPTY=true",
			}
		})
	result.Assert(c, icmd.Expected{ExitCode: 0})
	_, err := os.Stat(out)
	assert.Equal(c, true, os.IsNotExist(err))
}

func (s *BasicSuite) TestRoutesInputsToProperOutputsWithChmod(c *C) {
	oneOut := s.tmpDir.Join("one.out")
	twoOut := s.tmpDir.Join("two.out")
	result := icmd.RunCmd(icmd.Command(GomplateBin,
		"-f", s.tmpDir.Join("one"),
		"-f", s.tmpDir.Join("two"),
		"-o", oneOut,
		"-o", twoOut,
		"--chmod", "0600"), func(cmd *icmd.Cmd) {
		cmd.Stdin = bytes.NewBufferString("hello world")
	})
	result.Assert(c, icmd.Success)

	testdata := []struct {
		path    string
		mode    os.FileMode
		content string
	}{
		{oneOut, 0600, "hi\n"},
		{twoOut, 0600, "hello\n"},
	}
	for _, v := range testdata {
		info, err := os.Stat(v.path)
		assert.NilError(c, err)
		assert.Equal(c, config.NormalizeFileMode(v.mode), info.Mode())
		content, err := ioutil.ReadFile(v.path)
		assert.NilError(c, err)
		assert.Equal(c, v.content, string(content))
	}
}

func (s *BasicSuite) TestOverridesOutputModeWithChmod(c *C) {
	out := s.tmpDir.Join("two")
	result := icmd.RunCmd(icmd.Command(GomplateBin,
		"-f", s.tmpDir.Join("one"),
		"-o", out,
		"--chmod", "0600"), func(cmd *icmd.Cmd) {
		cmd.Stdin = bytes.NewBufferString("hello world")
	})
	result.Assert(c, icmd.Success)

	testdata := []struct {
		path    string
		mode    os.FileMode
		content string
	}{
		{out, 0600, "hi\n"},
	}
	for _, v := range testdata {
		info, err := os.Stat(v.path)
		assert.NilError(c, err)
		assert.Equal(c, config.NormalizeFileMode(v.mode), info.Mode())
		content, err := ioutil.ReadFile(v.path)
		assert.NilError(c, err)
		assert.Equal(c, v.content, string(content))
	}
}

func (s *BasicSuite) TestAppliesChmodBeforeWrite(c *C) {
	// 'broken' was created with mode 0000
	out := s.tmpDir.Join("broken")
	result := icmd.RunCmd(icmd.Command(GomplateBin,
		"-f", s.tmpDir.Join("one"),
		"-o", out,
		"--chmod", "0644"), func(cmd *icmd.Cmd) {
	})
	result.Assert(c, icmd.Success)

	info, err := os.Stat(out)
	assert.NilError(c, err)
	assert.Equal(c, config.NormalizeFileMode(0644), info.Mode())
	content, err := ioutil.ReadFile(out)
	assert.NilError(c, err)
	assert.Equal(c, "hi\n", string(content))
}
