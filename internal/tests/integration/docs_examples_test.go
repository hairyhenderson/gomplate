//+build integration

package integration

import (
	. "gopkg.in/check.v1"

	"gotest.tools/v3/fs"
	"gotest.tools/v3/icmd"
)

// This suite contains integration tests to make sure that (some of) the examples
// in the gomplate docs work correctly
type DocExamplesSuite struct {
	tmpDir *fs.Dir
}

var _ = Suite(&DocExamplesSuite{})

func (s *DocExamplesSuite) SetUpSuite(c *C) {
	s.tmpDir = fs.NewDir(c, "gomplate-inttests")
}

func (s *DocExamplesSuite) TearDownSuite(c *C) {
	s.tmpDir.Remove()
}

func (s *DocExamplesSuite) TestDataExamples(c *C) {
	result := icmd.RunCommand(GomplateBin,
		"-i", "{{ $rows := (jsonArray `[[\"first\",\"second\"],[\"1\",\"2\"],[\"3\",\"4\"]]`) }}{{ data.ToCSV \";\" $rows }}",
	)
	expected := "first;second\r\n1;2\r\n3;4\r\n"
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: expected})
}
