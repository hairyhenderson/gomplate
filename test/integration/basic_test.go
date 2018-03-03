//+build integration
//+build !windows

package integration

import (
	"bytes"
	"io/ioutil"

	. "gopkg.in/check.v1"

	"github.com/gotestyourself/gotestyourself/assert"
	"github.com/gotestyourself/gotestyourself/assert/cmp"
	"github.com/gotestyourself/gotestyourself/fs"
	"github.com/gotestyourself/gotestyourself/icmd"
)

type BasicSuite struct {
	tmpDir *fs.Dir
}

var _ = Suite(&BasicSuite{})

func (s *BasicSuite) SetUpSuite(c *C) {
	s.tmpDir = fs.NewDir(c, "gomplate-inttests",
		fs.WithFile("one", "hi\n"),
		fs.WithFile("two", "hello\n"))
}

func (s *BasicSuite) TearDownSuite(c *C) {
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
}

func (s *BasicSuite) TestTakesStdinWithFileFlag(c *C) {
	result := icmd.RunCmd(icmd.Command(GomplateBin, "--file", "-"), func(cmd *icmd.Cmd) {
		cmd.Stdin = bytes.NewBufferString("hello world")
	})
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "hello world"})
}
func (s *BasicSuite) TestWritesToStdoutWithOutFlag(c *C) {
	result := icmd.RunCmd(icmd.Command(GomplateBin, "--out", "-"), func(cmd *icmd.Cmd) {
		cmd.Stdin = bytes.NewBufferString("hello world")
	})
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "hello world"})
}

func (s *BasicSuite) TestIgnoresStdinWithInFlag(c *C) {
	result := icmd.RunCmd(icmd.Command(GomplateBin, "--in", "hi"), func(cmd *icmd.Cmd) {
		cmd.Stdin = bytes.NewBufferString("hello world")
	})
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "hi"})
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
		Err:      "Error: Must provide same number of --out (1) as --file (2) options",
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

	content, err := ioutil.ReadFile(oneOut)
	assert.NilError(c, err)
	assert.Equal(c, "hi\n", string(content))
	content, err = ioutil.ReadFile(twoOut)
	assert.NilError(c, err)
	assert.Equal(c, "hello\n", string(content))
}

func (s *BasicSuite) TestFlagRules(c *C) {
	result := icmd.RunCommand(GomplateBin, "-f", "-", "-i", "HELLO WORLD")
	result.Assert(c, icmd.Expected{
		ExitCode: 1,
		Out:      "--in and --file may not be used together",
	})

	result = icmd.RunCommand(GomplateBin, "--output-dir", ".")
	result.Assert(c, icmd.Expected{
		ExitCode: 1,
		Out:      "--input-dir must be set when --output-dir is set",
	})

	result = icmd.RunCommand(GomplateBin, "--input-dir", ".", "--in", "param")
	result.Assert(c, icmd.Expected{
		ExitCode: 1,
		Out:      "--input-dir can not be used together with --in or --file",
	})

	result = icmd.RunCommand(GomplateBin, "--input-dir", ".", "--file", "input.txt")
	result.Assert(c, icmd.Expected{
		ExitCode: 1,
		Out:      "--input-dir can not be used together with --in or --file",
	})

	result = icmd.RunCommand(GomplateBin, "--output-dir", ".", "--out", "param")
	result.Assert(c, icmd.Expected{
		ExitCode: 1,
		Out:      "--output-dir can not be used together with --out",
	})
}

func (s *BasicSuite) TestDelimsChangedThroughOpts(c *C) {
	result := icmd.RunCommand(GomplateBin,
		"--left-delim", "((",
		"--right-delim", "))",
		"-i", `((print "hi"))`)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "hi"})
}

func (s *BasicSuite) TestDelimsChangedThroughEnvVars(c *C) {
	result := icmd.RunCmd(icmd.Command(GomplateBin, "-i", `<<print "hi">>`),
		func(cmd *icmd.Cmd) {
			cmd.Env = []string{
				"GOMPLATE_LEFT_DELIM=<<",
				"GOMPLATE_RIGHT_DELIM=>>",
			}
		})
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "hi"})
}

func (s *BasicSuite) TestUnknownArgErrors(c *C) {
	result := icmd.RunCommand(GomplateBin, "-in", "flibbit")
	result.Assert(c, icmd.Expected{ExitCode: 1, Out: `unknown command "flibbit" for "gomplate"`})
}
