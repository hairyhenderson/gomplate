package datasource

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMimeType(t *testing.T) {
	s := &Data{URL: mustParseURL("http://example.com/list?type=a/b/c")}
	_, err := s.MediaType()
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
			s := &Data{URL: mustParseURL(d.url), MType: d.mediaType}
			mt, err := s.MediaType()
			assert.NoError(t, err)
			assert.Equal(t, d.expected, mt)
		})
	}
}

func TestMimeTypeWithArg(t *testing.T) {
	d := &Data{URL: mustParseURL("http://example.com"), Subpath: "h\nttp://foo"}
	_, err := d.MediaType()
	assert.Error(t, err)

	data := []struct {
		url       string
		mediaType string
		subpath   string
		expected  string
	}{
		{"http://example.com/unknown",
			"",
			"/foo.json",
			"application/json"},
		{"http://example.com/unknown",
			"",
			"foo.json",
			"application/json"},
		{"http://example.com/",
			"text/foo",
			"/foo.json",
			"text/foo"},
		{"git+https://example.com/myrepo",
			"",
			"//foo.yaml",
			"application/yaml"},
		{"http://example.com/foo.json",
			"",
			"/foo.yaml",
			"application/yaml"},
		{"http://example.com/foo.json?type=application/array+yaml",
			"",
			"/foo.yaml",
			"application/array+yaml"},
		{"http://example.com/foo.json?type=application/array+yaml",
			"",
			"/foo.yaml?type=application/yaml",
			"application/yaml"},
		{"http://example.com/foo.json?type=application/array+yaml",
			"text/plain",
			"/foo.yaml?type=application/yaml",
			"application/yaml"},
	}

	for i, d := range data {
		t.Run(fmt.Sprintf("%d:%q,%q,%q==%q", i, d.url, d.mediaType, d.subpath, d.expected), func(t *testing.T) {
			data := &Data{URL: mustParseURL(d.url), MType: d.mediaType, Subpath: d.subpath}
			mt, err := data.MediaType()
			assert.NoError(t, err)
			assert.Equal(t, d.expected, mt)
		})
	}
}
