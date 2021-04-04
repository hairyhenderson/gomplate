package datasources

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNoopURLBuilder(t *testing.T) {
	u := mustParseURL("foo")
	b := &noopURLBuilder{}
	actual, _ := b.BuildURL(u)
	assert.Same(t, u, actual)
}

func TestBoltDBURLBuilder(t *testing.T) {
	b := &boltDBURLBuilder{}
	_, err := b.BuildURL(mustParseURL("base"), "extra", "too many!")
	assert.Error(t, err)

	_, err = b.BuildURL(mustParseURL("base"))
	assert.Error(t, err)

	data := []struct {
		u        string
		arg      string
		expected string
	}{
		{"boltdb:///tmp/foo.db#Bucket1", "key1", "boltdb:///tmp/foo.db?key=key1#Bucket1"},
		{"boltdb:///tmp/foo.db?type=application/json#Bucket1", "key1", "boltdb:///tmp/foo.db?key=key1&type=application%2Fjson#Bucket1"},
		{"boltdb:///tmp/foo.db?type=text/csv#Bucket1", "key1?type=application/json", "boltdb:///tmp/foo.db?key=key1&type=application%2Fjson#Bucket1"},
	}
	for i, d := range data {
		d := d
		t.Run(fmt.Sprintf("%s_%d", d.u, i), func(t *testing.T) {
			u := mustParseURL(d.u)
			actual, err := b.BuildURL(u, d.arg)
			assert.NoError(t, err)
			assert.EqualValues(t, mustParseURL(d.expected), actual)
		})
	}
}

func TestHTTPURLBuilder(t *testing.T) {
	u := mustParseURL("foo")
	b := &httpURLBuilder{}
	actual, err := b.BuildURL(u)
	assert.NoError(t, err)
	assert.Same(t, u, actual)

	u = mustParseURL("https://example.com/foo.json")
	expected := mustParseURL("https://example.com/bar.yaml")
	actual, err = b.BuildURL(u, "/bar.yaml")
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)

	u = mustParseURL("https://example.com/foo.json?a=b")
	expected = mustParseURL("https://example.com/bar.yaml?a=b")
	actual, err = b.BuildURL(u, "/bar.yaml")
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)

	u = mustParseURL("https://example.com/foo.json?a=b")
	expected = mustParseURL("https://example.com/bar.yaml?a=b&c=d")
	actual, err = b.BuildURL(u, "/bar.yaml?c=d")
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)

	u = mustParseURL("https://example.com/?type=foo")
	expected = mustParseURL("https://example.com/bar.yaml?type=foo")
	actual, err = b.BuildURL(u, "bar.yaml")
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)

	// param in arg must override param in original
	u = mustParseURL("https://example.com/?a=b&c=d")
	expected = mustParseURL("https://example.com/bar.yaml?a=b&c=e")
	actual, err = b.BuildURL(u, "/bar.yaml?c=e")
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)

	u = mustParseURL("https://example.com/foo.json?a=b")
	expected = mustParseURL("https://example.com/bar.yaml?a=b&c=d#foo")
	actual, err = b.BuildURL(u, "/bar.yaml?c=d#foo")
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)

	u = mustParseURL("https://example.com/foo.json#baz")
	expected = mustParseURL("https://example.com/bar.yaml#baz")
	actual, err = b.BuildURL(u, "/bar.yaml")
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)

	_, err = b.BuildURL(u, "bar", "baz")
	assert.Error(t, err)

	_, err = b.BuildURL(u, "bar\t\n")
	assert.Error(t, err)
}

func TestFileURLBuilder(t *testing.T) {
	b := &httpURLBuilder{}
	_, err := b.BuildURL(mustParseURL("base"), "extra", "too many!")
	assert.Error(t, err)

	data := []struct {
		u        string
		arg      string
		expected string
	}{
		{"file:///foo", "", "file:///foo"},
		{"file:///tmp/foo.txt", "bar.txt", "file:///tmp/bar.txt"},
		{"file:///tmp/partial/", "", "file:///tmp/partial/"},
		{"file:///tmp/partial/?type=application/json", "", "file:///tmp/partial/?type=application/json"},
		{"file:///tmp/partial/?type=application/json", "foo.txt", "file:///tmp/partial/foo.txt?type=application/json"},
	}
	for i, d := range data {
		d := d
		t.Run(fmt.Sprintf("%s_%d", d.u, i), func(t *testing.T) {
			args := []string{d.arg}
			if d.arg == "" {
				args = nil
			}
			actual, err := b.BuildURL(mustParseURL(d.u), args...)
			assert.NoError(t, err)
			assert.EqualValues(t, mustParseURL(d.expected), actual)
		})
	}
}

func TestVaultURLBuilder(t *testing.T) {
	b := &vaultURLBuilder{}
	_, err := b.BuildURL(mustParseURL("base"), "extra", "too many!")
	assert.Error(t, err)

	data := []struct {
		u        *url.URL
		expected *url.URL
		args     []string
	}{
		{
			u:        mustParseURL("vault:///"),
			args:     []string{"ssh/creds/test?ip=10.1.2.3&username=user"},
			expected: mustParseURL("vault:///ssh/creds/test?ip=10.1.2.3&username=user"),
		},
	}

	for i, d := range data {
		d := d
		t.Run(fmt.Sprintf("%s_%d", d.u.String(), i), func(t *testing.T) {
			actual, err := b.BuildURL(d.u, d.args...)
			assert.NoError(t, err)
			assert.EqualValues(t, d.expected, actual)
		})
	}
}

func TestAWSSMURLBuilder(t *testing.T) {
	b := &awssmURLBuilder{}
	_, err := b.BuildURL(mustParseURL("base"), "extra", "too many!")
	assert.Error(t, err)

	data := []struct {
		u        *url.URL
		expected *url.URL
		args     []string
	}{
		{u: mustParseURL("noddy"), args: nil, expected: mustParseURL("noddy")},
		{u: mustParseURL("base"), args: []string{"extra"}, expected: mustParseURL("base/extra")},
		{u: mustParseURL("/foo/"), args: []string{"/extra"}, expected: mustParseURL("/foo/extra")},
		{u: mustParseURL("aws+sm:///foo"), args: []string{"bar"}, expected: mustParseURL("aws+sm:///foo/bar")},
		{u: mustParseURL("aws+sm:foo"), args: nil, expected: &url.URL{Scheme: "aws+sm", Path: "foo"}},
		{u: mustParseURL("aws+sm:foo/bar"), args: nil, expected: &url.URL{Scheme: "aws+sm", Path: "foo/bar"}},
		{u: mustParseURL("aws+sm:/foo/bar"), args: nil, expected: mustParseURL("aws+sm:///foo/bar")},
		{u: mustParseURL("aws+sm:foo"), args: []string{"baz"}, expected: &url.URL{Scheme: "aws+sm", Path: "foo/baz"}},
		{u: mustParseURL("aws+sm:foo/bar"), args: []string{"baz"}, expected: &url.URL{Scheme: "aws+sm", Path: "foo/bar/baz"}},
		{u: mustParseURL("aws+sm:/foo/bar"), args: []string{"baz"}, expected: mustParseURL("aws+sm:///foo/bar/baz")},
		{u: mustParseURL("aws+sm:///foo"), args: []string{"dir/"}, expected: mustParseURL("aws+sm:///foo/dir/")},
		{u: mustParseURL("aws+sm:///foo/"), args: nil, expected: mustParseURL("aws+sm:///foo/")},
		{u: mustParseURL("aws+sm:///foo/"), args: []string{"baz"}, expected: mustParseURL("aws+sm:///foo/baz")},

		{u: mustParseURL("aws+sm:foo?type=text/plain"), args: []string{"baz"},
			expected: &url.URL{Scheme: "aws+sm", Path: "foo/baz", RawQuery: "type=text/plain"}},
		{u: mustParseURL("aws+sm:foo/bar?type=text/plain"), args: []string{"baz"},
			expected: &url.URL{Scheme: "aws+sm", Path: "foo/bar/baz", RawQuery: "type=text/plain"}},
		{u: mustParseURL("aws+sm:/foo/bar?type=text/plain"), args: []string{"baz"},
			expected: &url.URL{Scheme: "aws+sm", Path: "/foo/bar/baz", RawQuery: "type=text/plain"}},
		{
			u:        mustParseURL("aws+sm:/foo/bar?type=text/plain"),
			args:     []string{"baz/qux?type=application/json&param=quux"},
			expected: mustParseURL("aws+sm:///foo/bar/baz/qux?param=quux&type=application%2Fjson"),
		},
	}

	for i, d := range data {
		d := d
		t.Run(fmt.Sprintf("%s_%d", d.u.String(), i), func(t *testing.T) {
			actual, err := b.BuildURL(d.u, d.args...)
			assert.NoError(t, err)
			assert.EqualValues(t, d.expected, actual)
		})
	}
}

func TestBlobBuilder(t *testing.T) {
	b := &blobURLBuilder{}
	_, err := b.BuildURL(mustParseURL("base"), "extra", "too many!")
	assert.Error(t, err)

	data := []struct {
		u        *url.URL
		expected *url.URL
		args     []string
	}{
		{u: mustParseURL("s3://mybucket/foo"), args: []string{"bar"}, expected: mustParseURL("s3://mybucket/foo/bar")},
		{u: mustParseURL("s3://mybucket/foo/bar"), expected: mustParseURL("s3://mybucket/foo/bar")},
		{u: mustParseURL("s3://mybucket/foo/bar"), args: []string{"baz"}, expected: mustParseURL("s3://mybucket/foo/bar/baz")},
		{u: mustParseURL("s3://mybucket/foo"), args: []string{"dir/"}, expected: mustParseURL("s3://mybucket/foo/dir/")},
		{u: mustParseURL("s3://mybucket/foo/"), expected: mustParseURL("s3://mybucket/foo/")},
		{u: mustParseURL("s3://mybucket/foo/"), args: []string{"baz"}, expected: mustParseURL("s3://mybucket/foo/baz")},
		{u: mustParseURL("s3://mybucket/foo?type=text/plain"), args: []string{"baz"},
			expected: &url.URL{Scheme: "s3", Host: "mybucket", Path: "/foo/baz", RawQuery: "type=text/plain"}},
		{u: mustParseURL("s3://mybucket/foo/bar?type=text/plain"), args: []string{"baz"},
			expected: &url.URL{Scheme: "s3", Host: "mybucket", Path: "/foo/bar/baz", RawQuery: "type=text/plain"}},
		{u: mustParseURL("s3://mybucket/foo/bar?type=text/plain"), args: []string{"baz"},
			expected: &url.URL{Scheme: "s3", Host: "mybucket", Path: "/foo/bar/baz", RawQuery: "type=text/plain"}},
		{
			u:        mustParseURL("s3://mybucket/dir1/?region=us-east-1&disableSSL=true&s3ForcePathStyle=true"),
			args:     []string{"?endpoint=example.com"},
			expected: mustParseURL("s3://mybucket/dir1/?disableSSL=true&endpoint=example.com&region=us-east-1&s3ForcePathStyle=true")},
		{
			u:        mustParseURL("s3://mybucket/foo/bar?type=text/plain"),
			args:     []string{"baz/qux?type=application/json&param=quux"},
			expected: mustParseURL("s3://mybucket/foo/bar/baz/qux?param=quux&type=application%2Fjson"),
		},
	}

	for i, d := range data {
		d := d
		t.Run(fmt.Sprintf("%s_%d", d.u.String(), i), func(t *testing.T) {
			actual, err := b.BuildURL(d.u, d.args...)
			assert.NoError(t, err)
			assert.EqualValues(t, d.expected, actual)
		})
	}
}

func TestGitParseArgPath(t *testing.T) {
	t.Parallel()
	g := &gitURLBuilder{}

	data := []struct {
		url        string
		arg        string
		repo, path string
	}{
		{"git+file:///foo//foo", "/bar", "", "/bar"},
		{"git+file:///foo//bar", "/baz//qux", "", "/baz//qux"},
		{"git+https://example.com/foo", "/bar", "/bar", ""},
		{"git+https://example.com/foo", "//bar", "", "//bar"},
		{"git+https://example.com/foo//bar", "//baz", "", "//baz"},
		{"git+https://example.com/foo", "/bar//baz", "/bar", "/baz"},
		{"git+https://example.com/foo?type=t", "/bar//baz", "/bar", "/baz"},
		{"git+https://example.com/foo#master", "/bar//baz", "/bar", "/baz"},
		{"git+https://example.com/foo", "//bar", "", "//bar"},
		{"git+https://example.com/foo?type=t", "//baz", "", "//baz"},
		{"git+https://example.com/foo?type=t#v1", "//bar", "", "//bar"},
	}

	for i, d := range data {
		d := d
		t.Run(fmt.Sprintf("%d:(%q,%q)==(%q,%q)", i, d.url, d.arg, d.repo, d.path), func(t *testing.T) {
			t.Parallel()
			u := mustParseURL(d.url)
			orig := u.Path
			repo, subpath := g.parseArgPath(orig, d.arg)
			assert.Equal(t, d.repo, repo)
			assert.Equal(t, d.path, subpath)
		})
	}
}

func TestGitURLBuilder(t *testing.T) {
	t.Parallel()
	b := &gitURLBuilder{}

	data := []struct {
		u        *url.URL
		expected *url.URL
		args     []string
	}{
		{u: mustParseURL("git+https://github.com/hairyhenderson/gomplate//docs-src/content/functions/aws.yml")},
		{u: mustParseURL("git+ssh://github.com/hairyhenderson/gomplate.git")},
		{u: mustParseURL("https://github.com")},
		{u: mustParseURL("git://example.com/foo//file.txt#someref")},
		{u: mustParseURL("git+file:///home/foo/repo//file.txt#someref")},
		{u: mustParseURL("git+file:///repo")},
		{u: mustParseURL("git+file:///foo//foo")},
		{
			u:        mustParseURL("git+file:///foo//foo"),
			args:     []string{"/bar"},
			expected: mustParseURL("git+file:///foo//foo/bar"),
		},
		{
			u: mustParseURL("git+file:///foo//bar"),
			// in this case the // is meaningless
			args:     []string{"/baz//qux"},
			expected: mustParseURL("git+file:///foo//bar/baz/qux"),
		},
		{
			u:        mustParseURL("git+https://example.com/foo"),
			args:     []string{"/bar"},
			expected: mustParseURL("git+https://example.com/foo/bar"),
		},
		{
			u:        mustParseURL("git+https://example.com/foo"),
			args:     []string{"//bar"},
			expected: mustParseURL("git+https://example.com/foo//bar"),
		},
		{
			u:        mustParseURL("git+https://example.com//foo"),
			args:     []string{"/bar"},
			expected: mustParseURL("git+https://example.com//foo/bar"),
		},
		{
			u:        mustParseURL("git+https://example.com/foo//bar"),
			args:     []string{"//baz"},
			expected: mustParseURL("git+https://example.com/foo//bar/baz"),
		},
		{
			u:        mustParseURL("git+https://example.com/foo"),
			args:     []string{"/bar//baz"},
			expected: mustParseURL("git+https://example.com/foo/bar//baz"),
		},
		{
			u:        mustParseURL("git+https://example.com/foo?type=t"),
			args:     []string{"/bar//baz"},
			expected: mustParseURL("git+https://example.com/foo/bar//baz?type=t"),
		},
		{
			u:        mustParseURL("git+https://example.com/foo#master"),
			args:     []string{"/bar//baz"},
			expected: mustParseURL("git+https://example.com/foo/bar//baz#master"),
		},
		{
			u:        mustParseURL("git+https://example.com/foo"),
			args:     []string{"/bar//baz?type=t"},
			expected: mustParseURL("git+https://example.com/foo/bar//baz?type=t"),
		},
		{
			u:        mustParseURL("git+https://example.com/foo"),
			args:     []string{"/bar//baz#master"},
			expected: mustParseURL("git+https://example.com/foo/bar//baz#master"),
		},
		{
			u:        mustParseURL("git+https://example.com/foo"),
			args:     []string{"//bar?type=t"},
			expected: mustParseURL("git+https://example.com/foo//bar?type=t"),
		},
		{
			u:        mustParseURL("git+https://example.com/foo"),
			args:     []string{"//bar#master"},
			expected: mustParseURL("git+https://example.com/foo//bar#master"),
		},
		{
			u:        mustParseURL("git+https://example.com/foo?type=t"),
			args:     []string{"//bar#master"},
			expected: mustParseURL("git+https://example.com/foo//bar?type=t#master"),
		},
		{
			u:        mustParseURL("git+https://example.com/foo?type=t#v1"),
			args:     []string{"//bar?type=j#v2"},
			expected: mustParseURL("git+https://example.com/foo//bar?type=t&type=j#v2"),
		},
	}

	for i, d := range data {
		d := d
		t.Run(fmt.Sprintf("%d:(%q,%q)==(%q)", i, d.u, d.args, d.expected), func(t *testing.T) {
			t.Parallel()

			if d.expected == nil {
				d.expected = d.u
			}

			out, err := b.BuildURL(d.u, d.args...)
			assert.NoError(t, err)
			assert.EqualValues(t, d.expected, out)
		})
	}
}
