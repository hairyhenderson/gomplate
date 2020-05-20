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
		if body == "" {
			// mirror back the headers
			h, err := json.Marshal(r.Header)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(code)
			fmt.Fprintln(w, string(h))
		} else {
			w.WriteHeader(code)
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

	expected := "{\"hello\": \"world\"}\n"

	ctx := context.Background()
	h := &HTTP{hc: client}

	data, err := h.Read(ctx, mustParseURL(server.URL))
	assert.NoError(t, err)
	assert.Equal(t, expected, string(data.Bytes))
}

func TestHTTPFileWithHeaders(t *testing.T) {
	server, client := setupHTTP(200, jsonMimetype, "")
	defer server.Close()

	hdr := http.Header{
		"Foo":             {"bar"},
		"foo":             {"baz"},
		"User-Agent":      {},
		"Accept-Encoding": {"test"},
	}
	h := &HTTP{hc: client}

	expected := http.Header{
		"Accept-Encoding": {"test"},
		"Foo":             {"bar", "baz"},
	}
	data, err := h.Read(
		context.WithValue(context.Background(), headerKey, hdr),
		mustParseURL("http://example.com/foo"))
	assert.NoError(t, err)
	actual := http.Header{}
	err = json.Unmarshal(data.Bytes, &actual)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)

	expected = http.Header{
		"Accept-Encoding": {"test"},
		"Foo":             {"bar", "baz"},
		"User-Agent":      {"Go-http-client/1.1"},
	}

	data, err = h.Read(
		context.WithValue(context.Background(), headerKey, expected),
		mustParseURL(server.URL))
	assert.NoError(t, err)
	actual = http.Header{}
	assert.NoError(t, err)
	err = json.Unmarshal(data.Bytes, &actual)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
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
