package datasources

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

func setupHTTP(code int, mimetype string, body string) (*httptest.Server, *http.Client) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", mimetype)
		w.WriteHeader(code)
		if body == "" {
			// mirror back the headers
			b, _ := json.Marshal(r.Header)
			fmt.Fprintln(w, string(b))
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

func assertJSONEqual(t *testing.T, expected, actual interface{}) {
	e, err := json.Marshal(expected)
	assert.NoError(t, err)
	a, err := json.Marshal(actual)
	assert.NoError(t, err)
	assert.Equal(t, string(e), string(a))
}

func TestHTTPRequester(t *testing.T) {
	server, client := setupHTTP(200, "application/json; charset=utf-8", `{"hello": "world"}`)
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r := &httpRequester{client}

	expected := map[string]interface{}{
		"hello": "world",
	}

	u := mustParseURL("http://example.com/foo")
	resp, err := r.Request(ctx, u, nil)
	assert.NoError(t, err)

	actual, err := resp.Parse()
	assert.NoError(t, err)
	assertJSONEqual(t, expected, actual)

	u = mustParseURL(server.URL)
	resp, err = r.Request(ctx, u, nil)
	assert.NoError(t, err)

	actual, err = resp.Parse()
	assert.NoError(t, err)
	assertJSONEqual(t, expected, actual)
}

func TestHTTPRequesterWithHeaders(t *testing.T) {
	server, client := setupHTTP(200, jsonMimetype, "")
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r := &httpRequester{client}
	expected := http.Header{
		"Accept-Encoding": {"test"},
		"Foo":             {"bar", "baz"},
	}

	u := mustParseURL("http://example.com/foo")
	resp, err := r.Request(ctx, u, http.Header{
		"Foo":             {"bar"},
		"foo":             {"baz"},
		"User-Agent":      {},
		"Accept-Encoding": {"test"},
	})
	assert.NoError(t, err)

	actual, err := resp.Parse()
	assert.NoError(t, err)
	assertJSONEqual(t, expected, actual)

	expected = http.Header{
		"Accept-Encoding": {"test"},
		"Foo":             {"bar", "baz"},
		"User-Agent":      {"Go-http-client/1.1"},
	}
	u = mustParseURL(server.URL)
	resp, err = r.Request(ctx, u, expected)
	assert.NoError(t, err)

	actual, err = resp.Parse()
	assert.NoError(t, err)
	assertJSONEqual(t, expected, actual)
}
