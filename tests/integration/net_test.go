//+build integration

package integration

import (
	. "gopkg.in/check.v1"

	"gotest.tools/v3/icmd"
)

type NetSuite struct{}

var _ = Suite(&NetSuite{})

func (s *NetSuite) TestLookupIP(c *C) {
	result := icmd.RunCommand(GomplateBin, "-i", `{{ net.LookupIP "localhost" }}`)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "127.0.0.1"})
}
