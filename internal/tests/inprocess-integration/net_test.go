package integration

import (
	. "gopkg.in/check.v1"
)

type NetSuite struct{}

var _ = Suite(&NetSuite{})

func (s *NetSuite) TestLookupIP(c *C) {
	inOutTest(c, `{{ net.LookupIP "localhost" }}`, "127.0.0.1")
}
