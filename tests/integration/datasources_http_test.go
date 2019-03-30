//+build integration

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
	http.HandleFunc("/not.json", typeHandler("application/yaml", "value: notjson\n"))
	http.HandleFunc("/foo", typeHandler("application/json", `{"value": "json"}`))
	http.HandleFunc("/actually.json", typeHandler("", `{"value": "json"}`))
	http.HandleFunc("/bogus.csv", typeHandler("text/plain", `{"value": "json"}`))
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

	result = icmd.RunCommand(GomplateBin,
		"-H", "foo=Foo:bar",
		"-i", "{{defineDatasource `foo` `http://"+s.l.Addr().String()+"/`}}{{ index (ds `foo`).headers.Foo 0 }}")
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "bar"})

	result = icmd.RunCommand(GomplateBin,
		"-i", "{{ $d := ds `http://"+s.l.Addr().String()+"/`}}{{ index (index $d.headers `Accept-Encoding`) 0 }}")
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "gzip"})
}

func (s *DatasourcesHTTPSuite) TestTypeOverridePrecedence(c *C) {
	result := icmd.RunCommand(GomplateBin,
		"-d", "foo=http://"+s.l.Addr().String()+"/foo",
		"-i", "{{ (ds `foo`).value }}")
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "json"})

	result = icmd.RunCommand(GomplateBin,
		"-d", "foo=http://"+s.l.Addr().String()+"/not.json",
		"-i", "{{ (ds `foo`).value }}")
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "notjson"})

	result = icmd.RunCommand(GomplateBin,
		"-d", "foo=http://"+s.l.Addr().String()+"/actually.json",
		"-i", "{{ (ds `foo`).value }}")
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "json"})

	result = icmd.RunCommand(GomplateBin,
		"-d", "foo=http://"+s.l.Addr().String()+"/bogus.csv?type=application/json",
		"-i", "{{ (ds `foo`).value }}")
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "json"})
}

func (s *DatasourcesHTTPSuite) TestAppendQueryAfterSubPaths(c *C) {
	result := icmd.RunCommand(GomplateBin,
		"-d", "foo=http://"+s.l.Addr().String()+"/?type=application/json",
		"-i", "{{ (ds `foo` `bogus.csv`).value }}")
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "json"})
}
