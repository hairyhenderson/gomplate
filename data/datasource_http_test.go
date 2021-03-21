package data

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/hairyhenderson/gomplate/v3/internal/config"
	"github.com/stretchr/testify/assert"
)

func setupHTTP(t *testing.T, code int, mimetype string, body string) *httptest.Server {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	t.Cleanup(srv.Close)

	return srv
}

func assertJSONEqual(t *testing.T, expected, actual interface{}) {
	e, err := json.Marshal(expected)
	assert.NoError(t, err)
	a, err := json.Marshal(actual)
	assert.NoError(t, err)
	assert.Equal(t, string(e), string(a))
}

func TestHTTPFile(t *testing.T) {
	srv := setupHTTP(t, 200, "application/json; charset=utf-8", `{"hello": "world"}`)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	u, _ := url.Parse(srv.URL)
	data := &Data{
		ds: map[string]config.DataSource{
			"foo": {URL: u},
		},
		ctx: ctx,
	}

	expected := map[string]interface{}{
		"hello": "world",
	}

	actual, err := data.Datasource("foo")
	assert.NoError(t, err)
	assertJSONEqual(t, expected, actual)

	actual, err = data.Datasource(srv.URL)
	assert.NoError(t, err)
	assertJSONEqual(t, expected, actual)
}

func TestHTTPFileWithHeaders(t *testing.T) {
	srv := setupHTTP(t, 200, jsonMimetype, "")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sources := make(map[string]config.DataSource)

	u, _ := url.Parse(srv.URL)
	sources["foo"] = config.DataSource{
		URL: u,
		Header: http.Header{
			"Foo":             {"bar"},
			"foo":             {"baz"},
			"User-Agent":      {},
			"Accept-Encoding": {"test"},
		},
	}
	data := &Data{
		ds:  sources,
		ctx: ctx,
	}
	expected := http.Header{
		"Accept-Encoding": {"test"},
		"Foo":             {"bar", "baz"},
	}
	actual, err := data.Datasource("foo")
	assert.NoError(t, err)
	assertJSONEqual(t, expected, actual)

	expected = http.Header{
		"Accept-Encoding": {"test"},
		"Foo":             {"bar", "baz"},
		"User-Agent":      {"Go-http-client/1.1"},
	}
	data = &Data{
		ds:           sources,
		extraHeaders: map[string]http.Header{srv.URL: expected},
		ctx:          ctx,
	}
	actual, err = data.Datasource(srv.URL)
	assert.NoError(t, err)
	assertJSONEqual(t, expected, actual)
}
