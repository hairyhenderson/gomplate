//+build integration
//+build !windows

package integration

import (
	. "gopkg.in/check.v1"

	"github.com/gotestyourself/gotestyourself/fs"
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
	inOutTest(c, `{{ file.Read "`+s.tmpDir.Join("one")+`"}}`, "hi")
}
