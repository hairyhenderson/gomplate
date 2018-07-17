//+build integration
//+build !windows

package integration

import (
	"io/ioutil"

	. "gopkg.in/check.v1"

	"github.com/gotestyourself/gotestyourself/assert"
	"github.com/gotestyourself/gotestyourself/fs"
	"github.com/gotestyourself/gotestyourself/icmd"
	tassert "github.com/stretchr/testify/assert"
)

type InputDirSuite struct {
	tmpDir *fs.Dir
}

var _ = Suite(&InputDirSuite{})

func (s *InputDirSuite) SetUpTest(c *C) {
	s.tmpDir = fs.NewDir(c, "gomplate-inttests",
		fs.WithFile("config.yml", "one: eins\ntwo: deux\n"),
		fs.WithDir("in",
			fs.WithFile("eins.txt", `{{ (ds "config").one }}`),
			fs.WithDir("inner",
				fs.WithFile("deux.txt", `{{ (ds "config").two }}`),
			),
		),
		fs.WithDir("out"),
		fs.WithDir("bad_in",
			fs.WithFile("bad.tmpl", "{{end}}"),
		),
	)
}

func (s *InputDirSuite) TearDownTest(c *C) {
	s.tmpDir.Remove()
}

func (s *InputDirSuite) TestInputDir(c *C) {
	result := icmd.RunCommand(GomplateBin,
		"--input-dir", s.tmpDir.Join("in"),
		"--output-dir", s.tmpDir.Join("out"),
		"-d", "config="+s.tmpDir.Join("config.yml"),
	)
	result.Assert(c, icmd.Success)

	files, err := ioutil.ReadDir(s.tmpDir.Join("out"))
	assert.NilError(c, err)
	tassert.Len(c, files, 2)

	files, err = ioutil.ReadDir(s.tmpDir.Join("out", "inner"))
	assert.NilError(c, err)
	tassert.Len(c, files, 1)

	content, err := ioutil.ReadFile(s.tmpDir.Join("out", "eins.txt"))
	assert.NilError(c, err)
	assert.Equal(c, "eins", string(content))

	content, err = ioutil.ReadFile(s.tmpDir.Join("out", "inner", "deux.txt"))
	assert.NilError(c, err)
	assert.Equal(c, "deux", string(content))
}

func (s *InputDirSuite) TestDefaultOutputDir(c *C) {
	result := icmd.RunCmd(icmd.Command(GomplateBin,
		"--input-dir", s.tmpDir.Join("in"),
		"-d", "config="+s.tmpDir.Join("config.yml"),
	), func(c *icmd.Cmd) {
		c.Dir = s.tmpDir.Join("out")
	})
	result.Assert(c, icmd.Success)

	files, err := ioutil.ReadDir(s.tmpDir.Join("out"))
	assert.NilError(c, err)
	tassert.Len(c, files, 2)

	files, err = ioutil.ReadDir(s.tmpDir.Join("out", "inner"))
	assert.NilError(c, err)
	tassert.Len(c, files, 1)

	content, err := ioutil.ReadFile(s.tmpDir.Join("out", "eins.txt"))
	assert.NilError(c, err)
	assert.Equal(c, "eins", string(content))

	content, err = ioutil.ReadFile(s.tmpDir.Join("out", "inner", "deux.txt"))
	assert.NilError(c, err)
	assert.Equal(c, "deux", string(content))
}

func (s *InputDirSuite) TestReportsFilenameWithBadInputFile(c *C) {
	result := icmd.RunCommand(GomplateBin,
		"--input-dir", s.tmpDir.Join("bad_in"),
		"--output-dir", s.tmpDir.Join("out"),
		"-d", "config="+s.tmpDir.Join("config.yml"),
	)
	result.Assert(c, icmd.Expected{
		ExitCode: 1,
		Out:      "template: " + s.tmpDir.Join("bad_in", "bad.tmpl") + ":1: unexpected {{end}}",
	})
}
