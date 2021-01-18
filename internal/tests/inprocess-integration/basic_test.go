package integration

import (
	"io/ioutil"
	"os"

	"github.com/hairyhenderson/gomplate/v3/internal/config"
	. "gopkg.in/check.v1"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
	"gotest.tools/v3/fs"
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
	o, e, err := cmdTest(c, "-v")
	assertSuccess(c, o, e, err, "gomplate version 0.0.0\n")
}

func (s *BasicSuite) TestTakesStdinByDefault(c *C) {
	o, e, err := cmdWithStdin(c, nil, "hello world")
	assertSuccess(c, o, e, err, "hello world")
}

func (s *BasicSuite) TestTakesStdinWithFileFlag(c *C) {
	o, e, err := cmdWithStdin(c, []string{"--file", "-"}, "hello world")
	assertSuccess(c, o, e, err, "hello world")
}

func (s *BasicSuite) TestWritesToStdoutWithOutFlag(c *C) {
	o, e, err := cmdWithStdin(c, []string{"--out", "-"}, "hello world")
	assertSuccess(c, o, e, err, "hello world")
}

func (s *BasicSuite) TestIgnoresStdinWithInFlag(c *C) {
	o, e, err := cmdWithStdin(c, []string{"--in", "hi"}, "hello world")
	assertSuccess(c, o, e, err, "hi")
}

func (s *BasicSuite) TestErrorsWithInputOutputImbalance(c *C) {
	_, _, err := cmdTest(c,
		"-f", s.tmpDir.Join("one"),
		"-f", s.tmpDir.Join("two"),
		"-o", s.tmpDir.Join("out"),
	)
	assert.ErrorContains(c, err, "must provide same number of 'outputFiles' (1) as 'in' or 'inputFiles' (2) options")
}

func (s *BasicSuite) TestRoutesInputsToProperOutputs(c *C) {
	oneOut := s.tmpDir.Join("one.out")
	twoOut := s.tmpDir.Join("two.out")

	o, e, err := cmdTest(c,
		"-f", s.tmpDir.Join("one"),
		"-f", s.tmpDir.Join("two"),
		"-o", oneOut,
		"-o", twoOut,
	)
	assertSuccess(c, o, e, err, "")

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
	testdata := []struct {
		args   []string
		errmsg string
	}{
		{
			[]string{"-f", "-", "-i", "HELLO WORLD"},
			"only one of these options is supported at a time: 'in', 'inputFiles'",
		},
		{
			[]string{"--output-dir", "."},
			"these options must be set together: 'outputDir', 'inputDir'",
		},
		{
			[]string{"--input-dir", ".", "--in", "param"},
			"only one of these options is supported at a time: 'in', 'inputDir'",
		},
		{
			[]string{"--input-dir", ".", "--file", "input.txt"},
			"only one of these options is supported at a time: 'inputFiles', 'inputDir'",
		},
		{
			[]string{"--output-dir", ".", "--out", "param"},
			"only one of these options is supported at a time: 'outputFiles', 'outputDir'",
		},
		{
			[]string{"--output-map", ".", "--out", "param"},
			"only one of these options is supported at a time: 'outputFiles', 'outputMap'",
		},
	}

	for _, d := range testdata {
		_, _, err := cmdTest(c, d.args...)
		assert.ErrorContains(c, err, d.errmsg)
	}
}

func (s *BasicSuite) TestDelimsChangedThroughOpts(c *C) {
	o, e, err := cmdTest(c,
		"--left-delim", "((",
		"--right-delim", "))",
		"-i", `foo((print "hi"))`,
	)
	assertSuccess(c, o, e, err, "foohi")
}

func (s *BasicSuite) TestDelimsChangedThroughEnvVars(c *C) {
	o, e, err := cmdWithEnv(c, []string{"-i", `foo<<print "hi">>`}, map[string]string{
		"GOMPLATE_LEFT_DELIM":  "<<",
		"GOMPLATE_RIGHT_DELIM": ">>",
	})
	assertSuccess(c, o, e, err, "foohi")
}

func (s *BasicSuite) TestUnknownArgErrors(c *C) {
	_, e, err := cmdTest(c, "-in", "flibbit")
	assert.ErrorContains(c, err, "unknown command \"flibbit\" for \"gomplate\"")
	assert.Assert(c, cmp.Contains(e, "Error: unknown command \"flibbit\" for \"gomplate\"\n"))
}

func (s *BasicSuite) TestExecCommand(c *C) {
	out := s.tmpDir.Join("out")
	o, e, err := cmdTest(c, "-i", `{{print "hello world"}}`,
		"-o", out,
		"--", "cat", out)
	assertSuccess(c, o, e, err, "hello world")
}

func (s *BasicSuite) TestPostRunExecPipe(c *C) {
	o, e, err := cmdTest(c,
		"-i", `{{print "hello world"}}`,
		"--exec-pipe",
		"--", "tr", "a-z", "A-Z")
	assertSuccess(c, o, e, err, "HELLO WORLD")
}

func (s *BasicSuite) TestEmptyOutputSuppression(c *C) {
	out := s.tmpDir.Join("out")
	o, e, err := cmdWithEnv(c, []string{
		"-i",
		`{{print "\t  \n\n\r\n\t\t     \v\n"}}`,
		"-o", out,
	}, map[string]string{
		"GOMPLATE_SUPPRESS_EMPTY": "true",
	})
	assertSuccess(c, o, e, err, "")

	_, err = os.Stat(out)
	assert.Equal(c, true, os.IsNotExist(err))
}

func (s *BasicSuite) TestRoutesInputsToProperOutputsWithChmod(c *C) {
	oneOut := s.tmpDir.Join("one.out")
	twoOut := s.tmpDir.Join("two.out")
	o, e, err := cmdWithStdin(c, []string{
		"-f", s.tmpDir.Join("one"),
		"-f", s.tmpDir.Join("two"),
		"-o", oneOut,
		"-o", twoOut,
		"--chmod", "0600"}, "hello world")
	assertSuccess(c, o, e, err, "")

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

	o, e, err := cmdWithStdin(c, []string{
		"-f", s.tmpDir.Join("one"),
		"-o", out,
		"--chmod", "0600"}, "hello world")
	assertSuccess(c, o, e, err, "")

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
	_, _, err := cmdTest(c,
		"-f", s.tmpDir.Join("one"),
		"-o", out,
		"--chmod", "0644")
	assert.NilError(c, err)

	info, err := os.Stat(out)
	assert.NilError(c, err)
	assert.Equal(c, config.NormalizeFileMode(0644), info.Mode())
	content, err := ioutil.ReadFile(out)
	assert.NilError(c, err)
	assert.Equal(c, "hi\n", string(content))
}
