package data

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
		Ctx:     context.Background(),
		Sources: sources,
	}

	expected := map[string]interface{}{
		"hello": "world",
	}

	actual, err := data.Datasource("foo")
	require.NoError(t, err)
	assert.Equal(t, must(marshalObj(expected, json.Marshal)), must(marshalObj(actual, json.Marshal)))

	actual, err = data.Datasource(server.URL)
	require.NoError(t, err)
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
		Header: http.Header{
			"Foo":             {"bar"},
			"foo":             {"baz"},
			"User-Agent":      {},
			"Accept-Encoding": {"test"},
		},
	}
	data := &Data{
		Ctx:     context.Background(),
		Sources: sources,
	}
	expected := http.Header{
		"Accept-Encoding": {"test"},
		"Foo":             {"bar", "baz"},
	}
	actual, err := data.Datasource("foo")
	require.NoError(t, err)
	assert.Equal(t, must(marshalObj(expected, json.Marshal)), must(marshalObj(actual, json.Marshal)))

	expected = http.Header{
		"Accept-Encoding": {"test"},
		"Foo":             {"bar", "baz"},
		"User-Agent":      {"Go-http-client/1.1"},
	}
	data = &Data{
		Ctx:          context.Background(),
		Sources:      sources,
		ExtraHeaders: map[string]http.Header{server.URL: expected},
	}
	actual, err = data.Datasource(server.URL)
	require.NoError(t, err)
	assert.Equal(t, must(marshalObj(expected, json.Marshal)), must(marshalObj(actual, json.Marshal)))
}

func TestBuildURL(t *testing.T) {
	expected := "https://example.com/index.html"
	base := mustParseURL(expected)
	u, err := buildURL(base)
	require.NoError(t, err)
	assert.Equal(t, expected, u.String())

	expected = "https://example.com/index.html"
	base = mustParseURL("https://example.com")
	u, err = buildURL(base, "index.html")
	require.NoError(t, err)
	assert.Equal(t, expected, u.String())

	expected = "https://example.com/a/b/c/index.html"
	base = mustParseURL("https://example.com/a/")
	u, err = buildURL(base, "b/c/index.html")
	require.NoError(t, err)
	assert.Equal(t, expected, u.String())

	expected = "https://example.com/bar/baz/index.html"
	base = mustParseURL("https://example.com/foo")
	u, err = buildURL(base, "bar/baz/index.html")
	require.NoError(t, err)
	assert.Equal(t, expected, u.String())
}
