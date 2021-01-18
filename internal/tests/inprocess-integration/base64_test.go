package integration

import (
	"gopkg.in/check.v1"
)

type Base64Suite struct{}

var _ = check.Suite(&Base64Suite{})

func (s *Base64Suite) TestBase64Encode(c *check.C) {
	inOutTest(c, `{{ "foo" | base64.Encode }}`, "Zm9v")
}

func (s *Base64Suite) TestBase64Decode(c *check.C) {
	inOutTest(c, `{{ "Zm9v" | base64.Decode }}`, "foo")
}
