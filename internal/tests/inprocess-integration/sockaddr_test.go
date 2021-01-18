package integration

import (
	. "gopkg.in/check.v1"
)

type SockaddrSuite struct{}

var _ = Suite(&SockaddrSuite{})

func (s *SockaddrSuite) TestSockaddr(c *C) {
	inOutContains(c, `{{ range (sockaddr.GetAllInterfaces | sockaddr.Include "type" "ipv4") -}}
{{ . | sockaddr.Attr "address" }}
{{end}}`, "127.0.0.1")
}
