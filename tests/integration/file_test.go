//+build integration

package integration

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/gotestyourself/gotestyourself/assert"

	. "gopkg.in/check.v1"

	"github.com/gotestyourself/gotestyourself/fs"
	"github.com/gotestyourself/gotestyourself/icmd"
)

type FileSuite struct {
	tmpDir *fs.Dir
}

var _ = Suite(&FileSuite{})

func (s *FileSuite) SetUpSuite(c *C) {
	s.tmpDir = fs.NewDir(c, "gomplate-inttests",
		fs.WithFile("one", "hi\n"),
		fs.WithFile("two", "hello\n"))
}

func (s *FileSuite) TearDownSuite(c *C) {
	s.tmpDir.Remove()
}

func (s *FileSuite) TestReadsFile(c *C) {
	inOutTest(c, "{{ file.Read `"+s.tmpDir.Join("one")+"`}}", "hi")
}

func (s *FileSuite) TestWrite(c *C) {
	outDir := s.tmpDir.Join("writeOutput")
	os.MkdirAll(outDir, 0755)
	result := icmd.RunCmd(icmd.Command(GomplateBin,
		"-i", `{{ "hello world" | file.Write "./out" }}`,
	), func(cmd *icmd.Cmd) {
		cmd.Dir = outDir
	})
	result.Assert(c, icmd.Expected{ExitCode: 0})
	out, err := ioutil.ReadFile(filepath.Join(outDir, "out"))
	assert.NilError(c, err)
	assert.Equal(c, "hello world", string(out))
}
