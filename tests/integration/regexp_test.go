//+build integration

package integration

import (
	. "gopkg.in/check.v1"

	"gotest.tools/v3/icmd"
)

type RegexpSuite struct{}

var _ = Suite(&RegexpSuite{})

func (s *RegexpSuite) TestReplace(c *C) {
	result := icmd.RunCommand(GomplateBin, "-i",
		`{{ "1.2.3-59" | regexp.Replace "-([0-9]*)" ".$1" }}`)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "1.2.3.59"})
}
