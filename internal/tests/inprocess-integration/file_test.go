package integration

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"gotest.tools/v3/assert"

	. "gopkg.in/check.v1"

	"gotest.tools/v3/fs"
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
	inOutTest(c, "{{ file.Read `"+s.tmpDir.Join("one")+"`}}", "hi\n")
}

func (s *FileSuite) TestWrite(c *C) {
	outDir := s.tmpDir.Join("writeOutput")
	os.MkdirAll(outDir, 0755)
	o, e, err := cmdWithDir(c, outDir, "-i", `{{ "hello world" | file.Write "./out" }}`)
	assertSuccess(c, o, e, err, "")

	out, err := ioutil.ReadFile(filepath.Join(outDir, "out"))
	assert.NilError(c, err)
	assert.Equal(c, "hello world", string(out))
}
