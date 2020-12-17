package datasources

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func mustParseURL(in string) *url.URL {
	u, _ := url.Parse(in)
	return u
}

func TestMimeAlias(t *testing.T) {
	t.Parallel()
	data := []struct {
		in, out string
	}{
		{csvMimetype, csvMimetype},
		{yamlMimetype, yamlMimetype},
		{"application/x-yaml", yamlMimetype},
	}

	for _, d := range data {
		assert.Equal(t, d.out, mimeAlias(d.in))
	}
}

func TestMimeType(t *testing.T) {
	u := mustParseURL("http://example.com/list?type=a/b/c")
	_, err := mimeType(u, "")
	assert.Error(t, err)

	data := []struct {
		url       string
		mediaType string
		expected  string
	}{
		{"http://example.com/foo.json",
			"",
			jsonMimetype},
		{"http://example.com/foo.json",
			"text/foo",
			"text/foo"},
		{"http://example.com/foo.json?type=application/yaml",
			"text/foo",
			"application/yaml"},
		{"http://example.com/list?type=application/array%2Bjson",
			"text/foo",
			"application/array+json"},
		{"http://example.com/list?type=application/array+json",
			"",
			"application/array+json"},
		{"http://example.com/unknown",
			"",
			"text/plain"},
	}

	for i, d := range data {
		t.Run(fmt.Sprintf("%d:%q,%q==%q", i, d.url, d.mediaType, d.expected), func(t *testing.T) {
			u := mustParseURL(d.url)
			mt, err := mimeType(u, d.mediaType)
			assert.NoError(t, err)
			assert.Equal(t, d.expected, mt)
		})
	}
}
