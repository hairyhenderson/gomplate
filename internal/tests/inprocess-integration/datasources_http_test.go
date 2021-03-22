package integration

import (
	"net/http"
	"net/http/httptest"

	. "gopkg.in/check.v1"
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

func (s *DatasourcesHTTPSuite) TestHTTPDatasource(c *C) {
	o, e, err := cmdTest(c,
		"-d", "foo="+s.srv.URL+"/mirror",
		"-H", "foo=Foo:bar",
		"-i", "{{ index (ds `foo`).headers.Foo 0 }}")
	assertSuccess(c, o, e, err, "bar")

	o, e, err = cmdTest(c,
		"-H", "foo=Foo:bar",
		"-i", "{{defineDatasource `foo` `"+s.srv.URL+"/mirror`}}{{ index (ds `foo`).headers.Foo 0 }}")
	assertSuccess(c, o, e, err, "bar")

	o, e, err = cmdTest(c,
		"-i", "{{ $d := ds `"+s.srv.URL+"/mirror`}}{{ index (index $d.headers `Accept-Encoding`) 0 }}")
	assertSuccess(c, o, e, err, "gzip")
}

func (s *DatasourcesHTTPSuite) TestTypeOverridePrecedence(c *C) {
	o, e, err := cmdTest(c,
		"-d", "foo="+s.srv.URL+"/foo",
		"-i", "{{ (ds `foo`).value }}")
	assertSuccess(c, o, e, err, "json")

	o, e, err = cmdTest(c,
		"-d", "foo="+s.srv.URL+"/not.json",
		"-i", "{{ (ds `foo`).value }}")
	assertSuccess(c, o, e, err, "notjson")

	o, e, err = cmdTest(c,
		"-d", "foo="+s.srv.URL+"/actually.json",
		"-i", "{{ (ds `foo`).value }}")
	assertSuccess(c, o, e, err, "json")

	o, e, err = cmdTest(c,
		"-d", "foo="+s.srv.URL+"/bogus.csv?type=application/json",
		"-i", "{{ (ds `foo`).value }}")
	assertSuccess(c, o, e, err, "json")

	o, e, err = cmdTest(c,
		"-c", ".="+s.srv.URL+"/list?type=application/array+json",
		"-i", "{{ range . }}{{ . }}{{ end }}")
	assertSuccess(c, o, e, err, "12345")
}

func (s *DatasourcesHTTPSuite) TestAppendQueryAfterSubPaths(c *C) {
	o, e, err := cmdTest(c,
		"-d", "foo="+s.srv.URL+"/?type=application/json",
		"-i", "{{ (ds `foo` `bogus.csv`).value }}")
	assertSuccess(c, o, e, err, "json")
}
