//+build integration

package integration

import (
	. "gopkg.in/check.v1"

	"github.com/gotestyourself/gotestyourself/icmd"
)

type SockaddrSuite struct{}

var _ = Suite(&SockaddrSuite{})

func (s *SockaddrSuite) TestSockaddr(c *C) {
	result := icmd.RunCommand(GomplateBin, "-i",
		`{{ range (sockaddr.GetAllInterfaces | sockaddr.Include "type" "ipv4") -}}
{{ . | sockaddr.Attr "address" }}
{{end}}`)
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "127.0.0.1"})
}
