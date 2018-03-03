//+build integration
//+build !windows

package integration

import (
	"net"
	"net/http"

	. "gopkg.in/check.v1"

	"github.com/gotestyourself/gotestyourself/icmd"
)

type DatasourcesHTTPSuite struct {
	l *net.TCPListener
}

var _ = Suite(&DatasourcesHTTPSuite{})

func (s *DatasourcesHTTPSuite) SetUpSuite(c *C) {
	var err error
	s.l, err = net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP("127.0.0.1")})
	handle(c, err)

	http.HandleFunc("/", mirrorHandler)
	go http.Serve(s.l, nil)
}

func (s *DatasourcesHTTPSuite) TearDownSuite(c *C) {
	s.l.Close()
}

func (s *DatasourcesHTTPSuite) TestReportsVersion(c *C) {
	result := icmd.RunCommand(GomplateBin,
		"-d", "foo=http://"+s.l.Addr().String()+"/",
		"-H", "foo=Foo:bar",
		"-i", "{{ index (ds `foo`).headers.Foo 0 }}")
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "bar"})
}
