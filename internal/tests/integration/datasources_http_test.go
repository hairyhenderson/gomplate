package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupDatasourcesHTTPTest(t *testing.T) *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/mirror", mirrorHandler)
	mux.HandleFunc("/not.json", typeHandler("application/yaml", "value: notjson\n"))
	mux.HandleFunc("/foo", typeHandler("application/json", `{"value": "json"}`))
	mux.HandleFunc("/actually.json", typeHandler("", `{"value": "json"}`))
	mux.HandleFunc("/bogus.csv", typeHandler("text/plain", `{"value": "json"}`))
	mux.HandleFunc("/list", typeHandler("application/array+json", `[1, 2, 3, 4, 5]`))

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	return srv
}

func TestDatasources_HTTP(t *testing.T) {
	srv := setupDatasourcesHTTPTest(t)

	o, e, err := cmd(t,
		"-d", "foo="+srv.URL+"/mirror",
		"-H", "foo=Foo:bar",
		"-i", "{{ index (ds `foo`).headers.Foo 0 }}").run()
	assertSuccess(t, o, e, err, "bar")

	o, e, err = cmd(t,
		"-H", "foo=Foo:bar",
		"-i", "{{defineDatasource `foo` `"+srv.URL+"/mirror`}}{{ index (ds `foo`).headers.Foo 0 }}").run()
	assertSuccess(t, o, e, err, "bar")

	o, e, err = cmd(t,
		"-i", "{{ $d := ds `"+srv.URL+"/mirror`}}{{ index (index $d.headers `Accept-Encoding`) 0 }}").run()
	assertSuccess(t, o, e, err, "gzip")
}

func TestDatasources_HTTP_TypeOverridePrecedence(t *testing.T) {
	srv := setupDatasourcesHTTPTest(t)

	o, e, err := cmd(t,
		"-d", "foo="+srv.URL+"/foo",
		"-i", "{{ (ds `foo`).value }}").run()
	assertSuccess(t, o, e, err, "json")

	o, e, err = cmd(t,
		"-d", "foo="+srv.URL+"/not.json",
		"-i", "{{ (ds `foo`).value }}").run()
	assertSuccess(t, o, e, err, "notjson")

	o, e, err = cmd(t,
		"-d", "foo="+srv.URL+"/actually.json",
		"-i", "{{ (ds `foo`).value }}").run()
	assertSuccess(t, o, e, err, "json")

	o, e, err = cmd(t,
		"-d", "foo="+srv.URL+"/bogus.csv?type=application/json",
		"-i", "{{ (ds `foo`).value }}").run()
	assertSuccess(t, o, e, err, "json")

	o, e, err = cmd(t,
		"-c", ".="+srv.URL+"/list?type=application/array+json",
		"-i", "{{ range . }}{{ . }}{{ end }}").run()
	assertSuccess(t, o, e, err, "12345")
}

func TestDatasources_HTTP_AppendQueryAfterSubPaths(t *testing.T) {
	srv := setupDatasourcesHTTPTest(t)

	o, e, err := cmd(t,
		"-d", "foo="+srv.URL+"/?type=application/json",
		"-i", "{{ (ds `foo` `bogus.csv`).value }}").run()
	assertSuccess(t, o, e, err, "json")
}
