// +build !windows

package data

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/blang/vfs"
	"github.com/blang/vfs/memfs"
	"github.com/stretchr/testify/assert"
)

var spyLogFatalfMsg string

func restoreLogFatalf() {
	logFatalf = log.Fatalf
}

func mockLogFatalf(msg string, args ...interface{}) {
	spyLogFatalfMsg = msg
	panic(spyLogFatalfMsg)
}

func setupMockLogFatalf() {
	logFatalf = mockLogFatalf
	spyLogFatalfMsg = ""
}

func TestNewSource(t *testing.T) {
	s := NewSource("foo", &url.URL{
		Scheme: "file",
		Path:   "/foo.json",
	})
	assert.Equal(t, "application/json", s.Type)
	assert.Equal(t, ".json", s.Ext)

	s = NewSource("foo", &url.URL{
		Scheme: "http",
		Host:   "example.com",
		Path:   "/foo.json",
	})
	assert.Equal(t, "application/json", s.Type)
	assert.Equal(t, ".json", s.Ext)

	s = NewSource("foo", &url.URL{
		Scheme: "ftp",
		Host:   "example.com",
		Path:   "/foo.json",
	})
	assert.Equal(t, "application/json", s.Type)
	assert.Equal(t, ".json", s.Ext)
}

func TestNewData(t *testing.T) {
	d := NewData(nil, nil)
	assert.Len(t, d.Sources, 0)

	d = NewData([]string{"foo=http:///foo.json"}, nil)
	assert.Equal(t, "/foo.json", d.Sources["foo"].URL.Path)

	d = NewData([]string{"foo=http:///foo.json"}, []string{})
	assert.Equal(t, "/foo.json", d.Sources["foo"].URL.Path)
	assert.Empty(t, d.Sources["foo"].Header)

	d = NewData([]string{"foo=http:///foo.json"}, []string{"bar=Accept: blah"})
	assert.Equal(t, "/foo.json", d.Sources["foo"].URL.Path)
	assert.Empty(t, d.Sources["foo"].Header)

	d = NewData([]string{"foo=http:///foo.json"}, []string{"foo=Accept: blah"})
	assert.Equal(t, "/foo.json", d.Sources["foo"].URL.Path)
	assert.Equal(t, "blah", d.Sources["foo"].Header["Accept"][0])
}

func TestParseSourceNoAlias(t *testing.T) {
	s, err := ParseSource("foo.json")
	assert.NoError(t, err)
	assert.Equal(t, "foo", s.Alias)

	_, err = ParseSource("../foo.json")
	assert.Error(t, err)

	_, err = ParseSource("ftp://example.com/foo.yml")
	assert.Error(t, err)
}

func TestParseSourceWithAlias(t *testing.T) {
	s, err := ParseSource("data=foo.json")
	assert.NoError(t, err)
	assert.Equal(t, "data", s.Alias)
	assert.Equal(t, "file", s.URL.Scheme)
	assert.Equal(t, "application/json", s.Type)
	assert.True(t, s.URL.IsAbs())

	s, err = ParseSource("data=/otherdir/foo.json")
	assert.NoError(t, err)
	assert.Equal(t, "data", s.Alias)
	assert.Equal(t, "file", s.URL.Scheme)
	assert.True(t, s.URL.IsAbs())
	assert.Equal(t, "/otherdir/foo.json", s.URL.Path)

	s, err = ParseSource("data=sftp://example.com/blahblah/foo.json")
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
				Alias: "foo",
				URL:   &url.URL{Scheme: "file", Path: "/tmp/" + fname},
				Ext:   ext,
				Type:  mime,
				FS:    fs,
			},
		}
		return &Data{Sources: sources}
	}
	test := func(ext, mime string, contents []byte) {
		data := setup(ext, mime, contents)
		expected := map[string]interface{}{"hello": map[interface{}]interface{}{"cruel": "world"}}
		actual := data.Datasource("foo")
		assert.Equal(t, expected, actual)
	}

	test("json", "application/json", []byte(`{"hello":{"cruel":"world"}}`))
	test("yml", "application/yaml", []byte("hello:\n  cruel: world\n"))

	d := setup("", "text/plain", nil)
	defer restoreLogFatalf()
	setupMockLogFatalf()
	assert.Panics(t, func() {
		d.Datasource("foo")
	})
	assert.Contains(t, spyLogFatalfMsg, "No value found for")
}

func TestDatasourceExists(t *testing.T) {
	sources := map[string]*Source{
		"foo": {Alias: "foo"},
	}
	data := &Data{Sources: sources}
	assert.True(t, data.DatasourceExists("foo"))
	assert.False(t, data.DatasourceExists("bar"))
}

func setupHTTP(code int, mimetype string, body string) (*httptest.Server, *http.Client) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", mimetype)
		w.WriteHeader(code)
		if body == "" {
			// mirror back the headers
			fmt.Fprintln(w, marshalObj(r.Header, json.Marshal))
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
		HC: client,
	}
	data := &Data{
		Sources: sources,
	}
	expected := make(map[string]interface{})
	expected["hello"] = "world"
	actual := data.Datasource("foo").(map[string]interface{})
	assert.Equal(t, expected["hello"], actual["hello"])
}

func TestHTTPFileWithHeaders(t *testing.T) {
	server, client := setupHTTP(200, "application/json", "")
	defer server.Close()

	sources := make(map[string]*Source)
	sources["foo"] = &Source{
		Alias: "foo",
		URL: &url.URL{
			Scheme: "http",
			Host:   "example.com",
			Path:   "/foo",
		},
		HC: client,
		Header: http.Header{
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
	actual := data.Datasource("foo")
	assert.Equal(t, marshalObj(expected, json.Marshal), marshalObj(actual, json.Marshal))
}

func TestParseHeaderArgs(t *testing.T) {
	args := []string{
		"foo=Accept: application/json",
		"bar=Authorization: Bearer supersecret",
	}
	expected := map[string]http.Header{
		"foo": {
			"Accept": {"application/json"},
		},
		"bar": {
			"Authorization": {"Bearer supersecret"},
		},
	}
	assert.Equal(t, expected, parseHeaderArgs(args))

	defer restoreLogFatalf()
	setupMockLogFatalf()
	assert.Panics(t, func() {
		parseHeaderArgs([]string{"foo"})
	})

	defer restoreLogFatalf()
	setupMockLogFatalf()
	assert.Panics(t, func() {
		parseHeaderArgs([]string{"foo=bar"})
	})

	args = []string{
		"foo=Accept: application/json",
		"foo=Foo: bar",
		"foo=foo: baz",
		"foo=fOO: qux",
		"bar=Authorization: Bearer  supersecret",
	}
	expected = map[string]http.Header{
		"foo": {
			"Accept": {"application/json"},
			"Foo":    {"bar", "baz", "qux"},
		},
		"bar": {
			"Authorization": {"Bearer  supersecret"},
		},
	}
	assert.Equal(t, expected, parseHeaderArgs(args))
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
			Alias: "foo",
			URL:   &url.URL{Scheme: "file", Path: "/tmp/" + fname},
			Ext:   ext,
			Type:  "text/plain",
			FS:    fs,
		},
	}
	data := &Data{
		Sources: sources,
	}
	actual := data.Include("foo")
	assert.Equal(t, contents, actual)
}
