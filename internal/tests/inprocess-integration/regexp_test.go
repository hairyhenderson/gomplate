package integration

import (
	. "gopkg.in/check.v1"
)

type RegexpSuite struct{}

var _ = Suite(&RegexpSuite{})

func (s *RegexpSuite) TestReplace(c *C) {
	inOutTest(c, `{{ "1.2.3-59" | regexp.Replace "-([0-9]*)" ".$1" }}`, "1.2.3.59")
}

func (s *RegexpSuite) TestQuoteMeta(c *C) {
	inOutTest(c, "{{ regexp.QuoteMeta `foo{(\\` }}", `foo\{\(\\`)
}
