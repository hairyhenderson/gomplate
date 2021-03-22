//+build integration

package integration

import (
	"net/http"
	"net/http/httptest"

	. "gopkg.in/check.v1"

	"gotest.tools/v3/icmd"
)

type DatasourcesHTTPSuite struct {
	srv *httptest.Server
}

var _ = Suite(&DatasourcesHTTPSuite{})

func (s *DatasourcesHTTPSuite) SetUpSuite(c *C) {
	mux := http.NewServeMux()
	mux.HandleFunc("/mirror", mirrorHandler)
	mux.HandleFunc("/not.json", typeHandler("application/yaml", "value: notjson\n"))
	mux.HandleFunc("/foo", typeHandler("application/json", `{"value": "json"}`))
	mux.HandleFunc("/actually.json", typeHandler("", `{"value": "json"}`))
	mux.HandleFunc("/bogus.csv", typeHandler("text/plain", `{"value": "json"}`))
	mux.HandleFunc("/list", typeHandler("application/array+json", `[1, 2, 3, 4, 5]`))

	s.srv = httptest.NewServer(mux)
}

func (s *DatasourcesHTTPSuite) TearDownSuite(c *C) {
	s.srv.Close()
}

func (s *DatasourcesHTTPSuite) TestReportsVersion(c *C) {
	result := icmd.RunCommand(GomplateBin,
		"-d", "foo="+s.srv.URL+"/mirror",
		"-H", "foo=Foo:bar",
		"-i", "{{ index (ds `foo`).headers.Foo 0 }}")
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "bar"})

	result = icmd.RunCommand(GomplateBin,
		"-H", "foo=Foo:bar",
		"-i", "{{defineDatasource `foo` `"+s.srv.URL+"/mirror`}}{{ index (ds `foo`).headers.Foo 0 }}")
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "bar"})

	result = icmd.RunCommand(GomplateBin,
		"-i", "{{ $d := ds `"+s.srv.URL+"/mirror`}}{{ index (index $d.headers `Accept-Encoding`) 0 }}")
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "gzip"})
}

func (s *DatasourcesHTTPSuite) TestTypeOverridePrecedence(c *C) {
	result := icmd.RunCommand(GomplateBin,
		"-d", "foo="+s.srv.URL+"/foo",
		"-i", "{{ (ds `foo`).value }}")
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "json"})

	result = icmd.RunCommand(GomplateBin,
		"-d", "foo="+s.srv.URL+"/not.json",
		"-i", "{{ (ds `foo`).value }}")
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "notjson"})

	result = icmd.RunCommand(GomplateBin,
		"-d", "foo="+s.srv.URL+"/actually.json",
		"-i", "{{ (ds `foo`).value }}")
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "json"})

	result = icmd.RunCommand(GomplateBin,
		"-d", "foo="+s.srv.URL+"/bogus.csv?type=application/json",
		"-i", "{{ (ds `foo`).value }}")
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "json"})

	result = icmd.RunCommand(GomplateBin,
		"-c", ".="+s.srv.URL+"/list?type=application/array+json",
		"-i", "{{ range . }}{{ . }}{{ end }}")
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "12345"})
}

func (s *DatasourcesHTTPSuite) TestAppendQueryAfterSubPaths(c *C) {
	result := icmd.RunCommand(GomplateBin,
		"-d", "foo="+s.srv.URL+"/?type=application/json",
		"-i", "{{ (ds `foo` `bogus.csv`).value }}")
	result.Assert(c, icmd.Expected{ExitCode: 0, Out: "json"})
}
