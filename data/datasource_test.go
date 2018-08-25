// +build !windows

package data

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/blang/vfs"
	"github.com/blang/vfs/memfs"
	"github.com/stretchr/testify/assert"
)

func TestNewData(t *testing.T) {
	d, err := NewData(nil, nil)
	assert.NoError(t, err)
	assert.Len(t, d.Sources, 0)

	d, err = NewData([]string{"foo=http:///foo.json"}, nil)
	assert.NoError(t, err)
	assert.Equal(t, "/foo.json", d.Sources["foo"].URL.Path)

	d, err = NewData([]string{"foo=http:///foo.json"}, []string{})
	assert.NoError(t, err)
	assert.Equal(t, "/foo.json", d.Sources["foo"].URL.Path)
	assert.Empty(t, d.Sources["foo"].header)

	d, err = NewData([]string{"foo=http:///foo.json"}, []string{"bar=Accept: blah"})
	assert.NoError(t, err)
	assert.Equal(t, "/foo.json", d.Sources["foo"].URL.Path)
	assert.Empty(t, d.Sources["foo"].header)

	d, err = NewData([]string{"foo=http:///foo.json"}, []string{"foo=Accept: blah"})
	assert.NoError(t, err)
	assert.Equal(t, "/foo.json", d.Sources["foo"].URL.Path)
	assert.Equal(t, "blah", d.Sources["foo"].header["Accept"][0])
}

func TestParseSourceNoAlias(t *testing.T) {
	s, err := parseSource("foo.json")
	assert.NoError(t, err)
	assert.Equal(t, "foo", s.Alias)

	_, err = parseSource("../foo.json")
	assert.Error(t, err)

	_, err = parseSource("ftp://example.com/foo.yml")
	assert.Error(t, err)
}

func TestParseSourceWithAlias(t *testing.T) {
	s, err := parseSource("data=foo.json")
	assert.NoError(t, err)
	assert.Equal(t, "data", s.Alias)
	assert.Equal(t, "file", s.URL.Scheme)
	assert.True(t, s.URL.IsAbs())

	s, err = parseSource("data=/otherdir/foo.json")
	assert.NoError(t, err)
	assert.Equal(t, "data", s.Alias)
	assert.Equal(t, "file", s.URL.Scheme)
	assert.True(t, s.URL.IsAbs())
	assert.Equal(t, "/otherdir/foo.json", s.URL.Path)

	s, err = parseSource("data=sftp://example.com/blahblah/foo.json")
	assert.NoError(t, err)
	assert.Equal(t, "data", s.Alias)
	assert.Equal(t, "sftp", s.URL.Scheme)
	assert.True(t, s.URL.IsAbs())
	assert.Equal(t, "/blahblah/foo.json", s.URL.Path)
}

func TestDatasource(t *testing.T) {
	setup := func(ext, mime string, contents []byte) *Data {
		fname := "foo." + ext
		fs := memfs.Create()
		_ = fs.Mkdir("/tmp", 0777)
		f, _ := vfs.Create(fs, "/tmp/"+fname)
		_, _ = f.Write(contents)

		sources := map[string]*Source{
			"foo": {
				Alias:     "foo",
				URL:       &url.URL{Scheme: "file", Path: "/tmp/" + fname},
				mediaType: mime,
				fs:        fs,
			},
		}
		return &Data{Sources: sources}
	}
	test := func(ext, mime string, contents []byte) {
		data := setup(ext, mime, contents)
		expected := map[string]interface{}{"hello": map[interface{}]interface{}{"cruel": "world"}}
		actual, err := data.Datasource("foo")
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	}

	test("json", jsonMimetype, []byte(`{"hello":{"cruel":"world"}}`))
	test("yml", yamlMimetype, []byte("hello:\n  cruel: world\n"))

	d := setup("", textMimetype, nil)
	actual, err := d.Datasource("foo")
	assert.NoError(t, err)
	assert.Equal(t, "", actual)

	_, err = d.Datasource("bar")
	assert.Error(t, err)
}

func TestDatasourceReachable(t *testing.T) {
	fname := "foo.json"
	fs := memfs.Create()
	_ = fs.Mkdir("/tmp", 0777)
	f, _ := vfs.Create(fs, "/tmp/"+fname)
	_, _ = f.Write([]byte("{}"))

	sources := map[string]*Source{
		"foo": {
			Alias:     "foo",
			URL:       &url.URL{Scheme: "file", Path: "/tmp/" + fname},
			mediaType: jsonMimetype,
			fs:        fs,
		},
		"bar": {
			Alias: "bar",
			URL:   &url.URL{Scheme: "file", Path: "/bogus"},
			fs:    fs,
		},
	}
	data := &Data{Sources: sources}

	assert.True(t, data.DatasourceReachable("foo"))
	assert.False(t, data.DatasourceReachable("bar"))
}

func TestDatasourceExists(t *testing.T) {
	sources := map[string]*Source{
		"foo": {Alias: "foo"},
	}
	data := &Data{Sources: sources}
	assert.True(t, data.DatasourceExists("foo"))
	assert.False(t, data.DatasourceExists("bar"))
}

func must(r interface{}, err error) interface{} {
	if err != nil {
		panic(err)
	}
	return r
}

func setupHTTP(code int, mimetype string, body string) (*httptest.Server, *http.Client) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", mimetype)
		w.WriteHeader(code)
		if body == "" {
			// mirror back the headers
			fmt.Fprintln(w, must(marshalObj(r.Header, json.Marshal)))
		} else {
			fmt.Fprintln(w, body)
		}
	}))

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				return url.Parse(server.URL)
			},
		},
	}

	return server, client
}

func TestHTTPFile(t *testing.T) {
	server, client := setupHTTP(200, "application/json; charset=utf-8", `{"hello": "world"}`)
	defer server.Close()

	sources := make(map[string]*Source)
	sources["foo"] = &Source{
		Alias: "foo",
		URL: &url.URL{
			Scheme: "http",
			Host:   "example.com",
			Path:   "/foo",
		},
		hc: client,
	}
	data := &Data{
		Sources: sources,
	}

	expected := map[string]interface{}{
		"hello": "world",
	}

	actual, err := data.Datasource("foo")
	assert.NoError(t, err)
	assert.Equal(t, must(marshalObj(expected, json.Marshal)), must(marshalObj(actual, json.Marshal)))

	actual, err = data.Datasource(server.URL)
	assert.NoError(t, err)
	assert.Equal(t, must(marshalObj(expected, json.Marshal)), must(marshalObj(actual, json.Marshal)))
}

func TestHTTPFileWithHeaders(t *testing.T) {
	server, client := setupHTTP(200, jsonMimetype, "")
	defer server.Close()

	sources := make(map[string]*Source)
	sources["foo"] = &Source{
		Alias: "foo",
		URL: &url.URL{
			Scheme: "http",
			Host:   "example.com",
			Path:   "/foo",
		},
		hc: client,
		header: http.Header{
			"Foo":             {"bar"},
			"foo":             {"baz"},
			"User-Agent":      {},
			"Accept-Encoding": {"test"},
		},
	}
	data := &Data{
		Sources: sources,
	}
	expected := http.Header{
		"Accept-Encoding": {"test"},
		"Foo":             {"bar", "baz"},
	}
	actual, err := data.Datasource("foo")
	assert.NoError(t, err)
	assert.Equal(t, must(marshalObj(expected, json.Marshal)), must(marshalObj(actual, json.Marshal)))

	expected = http.Header{
		"Accept-Encoding": {"test"},
		"Foo":             {"bar", "baz"},
		"User-Agent":      {"Go-http-client/1.1"},
	}
	data = &Data{
		Sources:      sources,
		extraHeaders: map[string]http.Header{server.URL: expected},
	}
	actual, err = data.Datasource(server.URL)
	assert.NoError(t, err)
	assert.Equal(t, must(marshalObj(expected, json.Marshal)), must(marshalObj(actual, json.Marshal)))
}

func TestParseHeaderArgs(t *testing.T) {
	args := []string{
		"foo=Accept: application/json",
		"bar=Authorization: Bearer supersecret",
	}
	expected := map[string]http.Header{
		"foo": {
			"Accept": {jsonMimetype},
		},
		"bar": {
			"Authorization": {"Bearer supersecret"},
		},
	}
	parsed, err := parseHeaderArgs(args)
	assert.NoError(t, err)
	assert.Equal(t, expected, parsed)

	_, err = parseHeaderArgs([]string{"foo"})
	assert.Error(t, err)

	_, err = parseHeaderArgs([]string{"foo=bar"})
	assert.Error(t, err)

	args = []string{
		"foo=Accept: application/json",
		"foo=Foo: bar",
		"foo=foo: baz",
		"foo=fOO: qux",
		"bar=Authorization: Bearer  supersecret",
	}
	expected = map[string]http.Header{
		"foo": {
			"Accept": {jsonMimetype},
			"Foo":    {"bar", "baz", "qux"},
		},
		"bar": {
			"Authorization": {"Bearer  supersecret"},
		},
	}
	parsed, err = parseHeaderArgs(args)
	assert.NoError(t, err)
	assert.Equal(t, expected, parsed)
}

func TestInclude(t *testing.T) {
	ext := "txt"
	contents := "hello world"
	fname := "foo." + ext
	fs := memfs.Create()
	_ = fs.Mkdir("/tmp", 0777)
	f, _ := vfs.Create(fs, "/tmp/"+fname)
	_, _ = f.Write([]byte(contents))

	sources := map[string]*Source{
		"foo": {
			Alias:     "foo",
			URL:       &url.URL{Scheme: "file", Path: "/tmp/" + fname},
			mediaType: textMimetype,
			fs:        fs,
		},
	}
	data := &Data{
		Sources: sources,
	}
	actual, err := data.Include("foo")
	assert.NoError(t, err)
	assert.Equal(t, contents, actual)
}

type errorReader struct{}

func (e errorReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("error")
}

func TestReadStdin(t *testing.T) {
	defer func() {
		stdin = nil
	}()
	stdin = strings.NewReader("foo")
	out, err := readStdin(nil)
	assert.NoError(t, err)
	assert.Equal(t, []byte("foo"), out)

	stdin = errorReader{}
	_, err = readStdin(nil)
	assert.Error(t, err)
}

// nolint: megacheck
func TestDefineDatasource(t *testing.T) {
	d := &Data{}
	_, err := d.DefineDatasource("", "foo.json")
	assert.Error(t, err)

	d = &Data{}
	_, err = d.DefineDatasource("", "../foo.json")
	assert.Error(t, err)

	d = &Data{}
	_, err = d.DefineDatasource("", "ftp://example.com/foo.yml")
	assert.Error(t, err)

	d = &Data{}
	_, err = d.DefineDatasource("data", "foo.json")
	s := d.Sources["data"]
	assert.NoError(t, err)
	assert.Equal(t, "data", s.Alias)
	assert.Equal(t, "file", s.URL.Scheme)
	assert.True(t, s.URL.IsAbs())

	d = &Data{}
	_, err = d.DefineDatasource("data", "/otherdir/foo.json")
	s = d.Sources["data"]
	assert.NoError(t, err)
	assert.Equal(t, "data", s.Alias)
	assert.Equal(t, "file", s.URL.Scheme)
	assert.True(t, s.URL.IsAbs())
	assert.Equal(t, "/otherdir/foo.json", s.URL.Path)

	d = &Data{}
	_, err = d.DefineDatasource("data", "sftp://example.com/blahblah/foo.json")
	s = d.Sources["data"]
	assert.NoError(t, err)
	assert.Equal(t, "data", s.Alias)
	assert.Equal(t, "sftp", s.URL.Scheme)
	assert.True(t, s.URL.IsAbs())
	assert.Equal(t, "/blahblah/foo.json", s.URL.Path)

	d = &Data{
		Sources: map[string]*Source{
			"data": {Alias: "data"},
		},
	}
	_, err = d.DefineDatasource("data", "/otherdir/foo.json")
	s = d.Sources["data"]
	assert.NoError(t, err)
	assert.Equal(t, "data", s.Alias)
	assert.Nil(t, s.URL)
}

func TestMimeType(t *testing.T) {
	s := &Source{URL: mustParseURL("http://example.com/foo.json")}
	mt, err := s.mimeType()
	assert.NoError(t, err)
	assert.Equal(t, jsonMimetype, mt)

	s = &Source{URL: mustParseURL("http://example.com/foo.json"), mediaType: "text/foo"}
	mt, err = s.mimeType()
	assert.NoError(t, err)
	assert.Equal(t, "text/foo", mt)

	s = &Source{URL: mustParseURL("http://example.com/foo.json"), mediaType: "text/foo"}
	mt, err = s.mimeType()
	assert.NoError(t, err)
	assert.Equal(t, "text/foo", mt)

	s = &Source{URL: mustParseURL("http://example.com/foo.json?type=application/yaml"), mediaType: "text/foo"}
	mt, err = s.mimeType()
	assert.NoError(t, err)
	assert.Equal(t, "application/yaml", mt)
}
