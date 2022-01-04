package datafs

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitFSMuxURL(t *testing.T) {
	testdata := []struct {
		in   string
		url  string
		file string
	}{
		{
			"http://example.com/foo.json",
			"http://example.com/",
			"foo.json",
		},
		{
			"http://example.com/foo.json?type=application/array+yaml",
			"http://example.com/?type=application/array+yaml",
			"foo.json",
		},
		{
			"vault:///secret/a/b/c",
			"vault:///",
			"secret/a/b/c",
		},
		{
			"vault:///secret/a/b/",
			"vault:///",
			"secret/a/b",
		},
		{
			"s3://bucket/a/b/",
			"s3://bucket/",
			"a/b",
		},
		{
			"vault:///foo/bar",
			"vault:///",
			"foo/bar",
		},
		{
			"consul://myhost/foo/bar/baz?q=1",
			"consul://myhost/?q=1",
			"foo/bar/baz",
		},
		{
			"git+https://example.com/myrepo//foo.yaml",
			"git+https://example.com/myrepo",
			"foo.yaml",
		},
		{
			"git+https://example.com/myrepo//",
			"git+https://example.com/myrepo",
			".",
		},
		{
			// git repos are special - no double-slash means the root
			"git+https://example.com/myrepo",
			"git+https://example.com/myrepo",
			".",
		},
		{
			"git+ssh://git@github.com/hairyhenderson/go-which.git//a/b/c/d?q=1",
			"git+ssh://git@github.com/hairyhenderson/go-which.git?q=1",
			"a/b/c/d",
		},
		{
			"merge:file:///tmp/jsonfile.json",
			"merge:///",
			"file:///tmp/jsonfile.json",
		},
		{
			"merge:a|b",
			"merge:///",
			"a|b",
		},
		{
			"merge:a|b|c|d|e",
			"merge:///",
			"a|b|c|d|e",
		},
		{
			"merge:foo/bar/baz.json|qux",
			"merge:///",
			"foo/bar/baz.json|qux",
		},
		{
			"merge:vault:///foo/bar|foo|git+ssh://git@github.com/hairyhenderson/go-which.git//a/b/c/d",
			"merge:///",
			"vault:///foo/bar|foo|git+ssh://git@github.com/hairyhenderson/go-which.git//a/b/c/d",
		},
	}

	for _, d := range testdata {
		u, err := url.Parse(d.in)
		assert.NoError(t, err)
		url, file := SplitFSMuxURL(u)
		assert.Equal(t, d.url, url.String())
		assert.Equal(t, d.file, file)
	}
}
