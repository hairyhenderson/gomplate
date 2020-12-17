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
		u        *url.URL
		args     []string
		expected *url.URL
	}{
		{mustParseURL("boltdb:///tmp/foo.db#Bucket1"), []string{"key1"}, mustParseURL("boltdb:///tmp/foo.db?key=key1#Bucket1")},
		{mustParseURL("boltdb:///tmp/foo.db?type=application/json#Bucket1"), []string{"key1"}, mustParseURL("boltdb:///tmp/foo.db?key=key1&type=application%2Fjson#Bucket1")},
		{mustParseURL("boltdb:///tmp/foo.db?type=text/csv#Bucket1"), []string{"key1?type=application/json"}, mustParseURL("boltdb:///tmp/foo.db?key=key1&type=application%2Fjson#Bucket1")},
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
		u        *url.URL
		args     []string
		expected *url.URL
	}{
		{mustParseURL("file:///foo"), nil, mustParseURL("file:///foo")},
		{mustParseURL("file:///tmp/foo.txt"), []string{"bar.txt"}, mustParseURL("file:///tmp/bar.txt")},
		{mustParseURL("file:///tmp/partial/"), nil, mustParseURL("file:///tmp/partial/")},
		{mustParseURL("file:///tmp/partial/?type=application/json"), nil, mustParseURL("file:///tmp/partial/?type=application/json")},
		{mustParseURL("file:///tmp/partial/?type=application/json"), []string{"foo.txt"}, mustParseURL("file:///tmp/partial/foo.txt?type=application/json")},
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
		args     []string
		expected *url.URL
	}{
		{mustParseURL("noddy"), nil, mustParseURL("noddy")},
		{mustParseURL("base"), []string{"extra"}, mustParseURL("base/extra")},
		{mustParseURL("/foo/"), []string{"/extra"}, mustParseURL("/foo/extra")},
		{mustParseURL("aws+sm:///foo"), []string{"bar"}, mustParseURL("aws+sm:///foo/bar")},
		{mustParseURL("aws+sm:foo"), nil, &url.URL{Scheme: "aws+sm", Path: "foo"}},
		{mustParseURL("aws+sm:foo/bar"), nil, &url.URL{Scheme: "aws+sm", Path: "foo/bar"}},
		{mustParseURL("aws+sm:/foo/bar"), nil, mustParseURL("aws+sm:///foo/bar")},
		{mustParseURL("aws+sm:foo"), []string{"baz"}, &url.URL{Scheme: "aws+sm", Path: "foo/baz"}},
		{mustParseURL("aws+sm:foo/bar"), []string{"baz"}, &url.URL{Scheme: "aws+sm", Path: "foo/bar/baz"}},
		{mustParseURL("aws+sm:/foo/bar"), []string{"baz"}, mustParseURL("aws+sm:///foo/bar/baz")},
		{mustParseURL("aws+sm:///foo"), []string{"dir/"}, mustParseURL("aws+sm:///foo/dir/")},
		{mustParseURL("aws+sm:///foo/"), nil, mustParseURL("aws+sm:///foo/")},
		{mustParseURL("aws+sm:///foo/"), []string{"baz"}, mustParseURL("aws+sm:///foo/baz")},

		{mustParseURL("aws+sm:foo?type=text/plain"), []string{"baz"},
			&url.URL{Scheme: "aws+sm", Path: "foo/baz", RawQuery: "type=text/plain"}},
		{mustParseURL("aws+sm:foo/bar?type=text/plain"), []string{"baz"},
			&url.URL{Scheme: "aws+sm", Path: "foo/bar/baz", RawQuery: "type=text/plain"}},
		{mustParseURL("aws+sm:/foo/bar?type=text/plain"), []string{"baz"},
			&url.URL{Scheme: "aws+sm", Path: "/foo/bar/baz", RawQuery: "type=text/plain"}},
		{
			mustParseURL("aws+sm:/foo/bar?type=text/plain"),
			[]string{"baz/qux?type=application/json&param=quux"},
			mustParseURL("aws+sm:///foo/bar/baz/qux?param=quux&type=application%2Fjson"),
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
		args     []string
		expected *url.URL
	}{
		{mustParseURL("s3://mybucket/foo"), []string{"bar"}, mustParseURL("s3://mybucket/foo/bar")},
		{mustParseURL("s3://mybucket/foo/bar"), nil, mustParseURL("s3://mybucket/foo/bar")},
		{mustParseURL("s3://mybucket/foo/bar"), []string{"baz"}, mustParseURL("s3://mybucket/foo/bar/baz")},
		{mustParseURL("s3://mybucket/foo"), []string{"dir/"}, mustParseURL("s3://mybucket/foo/dir/")},

		{mustParseURL("s3://mybucket/foo/"), nil, mustParseURL("s3://mybucket/foo/")},
		{mustParseURL("s3://mybucket/foo/"), []string{"baz"}, mustParseURL("s3://mybucket/foo/baz")},
		{mustParseURL("s3://mybucket/foo?type=text/plain"), []string{"baz"},
			&url.URL{Scheme: "s3", Host: "mybucket", Path: "/foo/baz", RawQuery: "type=text/plain"}},
		{mustParseURL("s3://mybucket/foo/bar?type=text/plain"), []string{"baz"},
			&url.URL{Scheme: "s3", Host: "mybucket", Path: "/foo/bar/baz", RawQuery: "type=text/plain"}},

		{mustParseURL("s3://mybucket/foo/bar?type=text/plain"), []string{"baz"},
			&url.URL{Scheme: "s3", Host: "mybucket", Path: "/foo/bar/baz", RawQuery: "type=text/plain"}},
		{
			mustParseURL("s3://mybucket/dir1/?region=us-east-1&disableSSL=true&s3ForcePathStyle=true"),
			[]string{"?endpoint=example.com"},
			mustParseURL("s3://mybucket/dir1/?disableSSL=true&endpoint=example.com&region=us-east-1&s3ForcePathStyle=true")},
		{
			mustParseURL("s3://mybucket/foo/bar?type=text/plain"),
			[]string{"baz/qux?type=application/json&param=quux"},
			mustParseURL("s3://mybucket/foo/bar/baz/qux?param=quux&type=application%2Fjson"),
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
			u, _ := url.Parse(d.url)
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
		url      *url.URL
		args     []string
		expected *url.URL
	}{
		{
			mustParseURL("git+https://github.com/hairyhenderson/gomplate//docs-src/content/functions/aws.yml"),
			nil,
			mustParseURL("git+https://github.com/hairyhenderson/gomplate//docs-src/content/functions/aws.yml")},
		{
			mustParseURL("git+ssh://github.com/hairyhenderson/gomplate.git"),
			nil,
			mustParseURL("git+ssh://github.com/hairyhenderson/gomplate.git"),
		},
		{
			mustParseURL("https://github.com"),
			nil,
			mustParseURL("https://github.com"),
		},
		{
			mustParseURL("git://example.com/foo//file.txt#someref"),
			nil,
			mustParseURL("git://example.com/foo//file.txt#someref"),
		},
		{
			mustParseURL("git+file:///home/foo/repo//file.txt#someref"),
			nil,
			mustParseURL("git+file:///home/foo/repo//file.txt#someref"),
		},
		{
			mustParseURL("git+file:///repo"),
			nil,
			mustParseURL("git+file:///repo"),
		},
		{
			mustParseURL("git+file:///foo//foo"),
			nil,
			mustParseURL("git+file:///foo//foo"),
		},
		{
			mustParseURL("git+file:///foo//foo"),
			[]string{"/bar"},
			mustParseURL("git+file:///foo//foo/bar"),
		},
		{
			mustParseURL("git+file:///foo//bar"),
			// in this case the // is meaningless
			[]string{"/baz//qux"},
			mustParseURL("git+file:///foo//bar/baz/qux"),
		},
		{
			mustParseURL("git+https://example.com/foo"),
			[]string{"/bar"},
			mustParseURL("git+https://example.com/foo/bar"),
		},
		{
			mustParseURL("git+https://example.com/foo"),
			[]string{"//bar"},
			mustParseURL("git+https://example.com/foo//bar"),
		},
		{
			mustParseURL("git+https://example.com//foo"),
			[]string{"/bar"},
			mustParseURL("git+https://example.com//foo/bar"),
		},
		{
			mustParseURL("git+https://example.com/foo//bar"),
			[]string{"//baz"},
			mustParseURL("git+https://example.com/foo//bar/baz"),
		},
		{
			mustParseURL("git+https://example.com/foo"),
			[]string{"/bar//baz"},
			mustParseURL("git+https://example.com/foo/bar//baz"),
		},
		{
			mustParseURL("git+https://example.com/foo?type=t"),
			[]string{"/bar//baz"},
			mustParseURL("git+https://example.com/foo/bar//baz?type=t"),
		},
		{
			mustParseURL("git+https://example.com/foo#master"),
			[]string{"/bar//baz"},
			mustParseURL("git+https://example.com/foo/bar//baz#master"),
		},
		{
			mustParseURL("git+https://example.com/foo"),
			[]string{"/bar//baz?type=t"},
			mustParseURL("git+https://example.com/foo/bar//baz?type=t"),
		},
		{
			mustParseURL("git+https://example.com/foo"),
			[]string{"/bar//baz#master"},
			mustParseURL("git+https://example.com/foo/bar//baz#master"),
		},
		{
			mustParseURL("git+https://example.com/foo"),
			[]string{"//bar?type=t"},
			mustParseURL("git+https://example.com/foo//bar?type=t"),
		},
		{
			mustParseURL("git+https://example.com/foo"),
			[]string{"//bar#master"},
			mustParseURL("git+https://example.com/foo//bar#master"),
		},
		{
			mustParseURL("git+https://example.com/foo?type=t"),
			[]string{"//bar#master"},
			mustParseURL("git+https://example.com/foo//bar?type=t#master"),
		},
		{
			mustParseURL("git+https://example.com/foo?type=t#v1"),
			[]string{"//bar?type=j#v2"},
			mustParseURL("git+https://example.com/foo//bar?type=t&type=j#v2"),
		},
	}

	for i, d := range data {
		d := d
		t.Run(fmt.Sprintf("%d:(%q,%q)==(%q)", i, d.url, d.args, d.expected), func(t *testing.T) {
			t.Parallel()
			out, err := b.BuildURL(d.url, d.args...)
			assert.NoError(t, err)
			assert.EqualValues(t, d.expected, out)
		})
	}
}
