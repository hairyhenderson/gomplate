package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/blang/vfs"
	"github.com/blang/vfs/memfs"
	"github.com/stretchr/testify/assert"
)

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
	test := func(ext, mime, contents string) {
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
				Type:  mime,
				FS:    fs,
			},
		}
		data := &Data{
			Sources: sources,
		}
		expected := map[string]interface{}{"hello": "world"}
		actual := data.Datasource("foo")
		assert.Equal(t, expected["hello"], actual["hello"])
	}

	test("json", "application/json", `{"hello":"world"}`)
	test("yml", "application/yaml", `hello: world`)
}

func setupHTTP(code int, mimetype string, body string) (*httptest.Server, *http.Client) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", mimetype)
		w.WriteHeader(code)
		fmt.Fprintln(w, body)
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
	actual := data.Datasource("foo")
	assert.Equal(t, expected["hello"], actual["hello"])
}
