package datasource

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
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
			fmt.Fprintln(w, must(json.Marshal(r.Header)))
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

	expected := map[string]interface{}{
		"hello": "world",
	}

	ctx := context.Background()
	h := &HTTP{hc: client}

	data, err := h.Read(ctx, mustParseURL("http://example.com/foo"))
	assert.NoError(t, err)
	assert.Equal(t, must(json.Marshal(expected)), must(json.Marshal(data.Bytes)))

	data, err = h.Read(ctx, mustParseURL(server.URL))
	assert.NoError(t, err)
	assert.Equal(t, must(json.Marshal(expected)), must(json.Marshal(data.Bytes)))
}

func TestHTTPFileWithHeaders(t *testing.T) {
	server, client := setupHTTP(200, jsonMimetype, "")
	defer server.Close()

	// TODO: uncomment
	// hdr := http.Header{
	// 	"Foo":             {"bar"},
	// 	"foo":             {"baz"},
	// 	"User-Agent":      {},
	// 	"Accept-Encoding": {"test"},
	// }
	ctx := context.Background()
	h := &HTTP{hc: client}

	expected := http.Header{
		"Accept-Encoding": {"test"},
		"Foo":             {"bar", "baz"},
	}
	u := mustParseURL("http://example.com/foo")
	data, err := h.Read(ctx, u)
	assert.NoError(t, err)
	assert.Equal(t, must(json.Marshal(expected)), must(json.Marshal(data.Bytes)))

	expected = http.Header{
		"Accept-Encoding": {"test"},
		"Foo":             {"bar", "baz"},
		"User-Agent":      {"Go-http-client/1.1"},
	}

	// TODO: uncomment
	// extraHeaders := http.Header{server.URL: expected}
	data, err = h.Read(ctx, mustParseURL(server.URL))
	assert.NoError(t, err)
	assert.Equal(t, must(json.Marshal(expected)), must(json.Marshal(data.Bytes)))
}

func TestBuildURL(t *testing.T) {
	expected := "https://example.com/index.html"
	base := mustParseURL(expected)
	u, err := buildURL(base)
	assert.NoError(t, err)
	assert.Equal(t, expected, u.String())

	expected = "https://example.com/index.html"
	base = mustParseURL("https://example.com")
	u, err = buildURL(base, "index.html")
	assert.NoError(t, err)
	assert.Equal(t, expected, u.String())

	expected = "https://example.com/a/b/c/index.html"
	base = mustParseURL("https://example.com/a/")
	u, err = buildURL(base, "b/c/index.html")
	assert.NoError(t, err)
	assert.Equal(t, expected, u.String())

	expected = "https://example.com/bar/baz/index.html"
	base = mustParseURL("https://example.com/foo")
	u, err = buildURL(base, "bar/baz/index.html")
	assert.NoError(t, err)
	assert.Equal(t, expected, u.String())
}
