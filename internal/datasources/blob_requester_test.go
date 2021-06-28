package datasources

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/johannesboyne/gofakes3"
	"github.com/johannesboyne/gofakes3/backend/s3mem"
	"github.com/stretchr/testify/assert"
)

func setupTestBucket(t *testing.T) (*httptest.Server, *url.URL) {
	backend := s3mem.New()
	faker := gofakes3.New(backend)
	ts := httptest.NewServer(faker.Server())

	err := backend.CreateBucket("mybucket")
	assert.NoError(t, err)
	c := "hello"
	err = putFile(backend, "mybucket", "file1", "text/plain", c)
	assert.NoError(t, err)

	c = `{"value": "goodbye world"}`
	err = putFile(backend, "mybucket", "file2", "application/json", c)
	assert.NoError(t, err)

	c = `value: what a world`
	err = putFile(backend, "mybucket", "file3", "application/yaml", c)
	assert.NoError(t, err)

	c = `value: out of this world`
	err = putFile(backend, "mybucket", "dir1/file1", "application/yaml", c)
	assert.NoError(t, err)

	c = `value: foo`
	err = putFile(backend, "mybucket", "dir1/file2", "application/yaml", c)
	assert.NoError(t, err)

	u, _ := url.Parse(ts.URL)
	return ts, u
}

func putFile(backend gofakes3.Backend, bucket, file, mime, content string) error {
	_, err := backend.PutObject(
		bucket,
		file,
		map[string]string{"Content-Type": mime},
		bytes.NewBufferString(content),
		int64(len(content)),
	)
	return err
}

func TestReadBlob(t *testing.T) {
	ctx := context.Background()

	ts, u := setupTestBucket(t)
	defer ts.Close()

	os.Setenv("AWS_ANON", "true")
	defer os.Unsetenv("AWS_ANON")

	r := &blobRequester{}

	var expected interface{}
	expected = "hello"
	out, err := r.Request(ctx, mustParseURL("s3://mybucket/file1?region=us-east-1&disableSSL=true&s3ForcePathStyle=true&type=text/plain&endpoint="+u.Host), nil)
	assert.NoError(t, err)
	defer out.Body.Close()
	actual, err := ioutil.ReadAll(out.Body)
	assert.NoError(t, err)
	assert.Equal(t, expected, string(actual))

	os.Unsetenv("AWS_ANON")

	os.Setenv("AWS_ACCESS_KEY_ID", "fake")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "fake")
	defer os.Unsetenv("AWS_ACCESS_KEY_ID")
	defer os.Unsetenv("AWS_SECRET_ACCESS_KEY")
	os.Setenv("AWS_S3_ENDPOINT", u.Host)
	defer os.Unsetenv("AWS_S3_ENDPOINT")

	expected = `{"value": "goodbye world"}`
	out, err = r.Request(ctx, mustParseURL("s3://mybucket/file2?region=us-east-1&disableSSL=true&s3ForcePathStyle=true"), nil)
	assert.NoError(t, err)
	defer out.Body.Close()
	actual, err = ioutil.ReadAll(out.Body)
	assert.NoError(t, err)
	assert.Equal(t, expected, string(actual))

	expected = `["dir1/","file1","file2","file3"]`
	out, err = r.Request(ctx, mustParseURL("s3://mybucket/?region=us-east-1&disableSSL=true&s3ForcePathStyle=true"), nil)
	assert.NoError(t, err)
	defer out.Body.Close()
	actual, err = ioutil.ReadAll(out.Body)
	assert.NoError(t, err)
	assert.EqualValues(t, expected, string(actual))

	expected = `["file1","file2"]`
	out, err = r.Request(ctx, mustParseURL("s3://mybucket/dir1/?region=us-east-1&disableSSL=true&s3ForcePathStyle=true"), nil)
	assert.NoError(t, err)
	defer out.Body.Close()
	actual, err = ioutil.ReadAll(out.Body)
	assert.NoError(t, err)
	assert.EqualValues(t, expected, string(actual))
}

func TestBlobSanitizeURL(t *testing.T) {
	data := []struct {
		in       string
		expected string
	}{
		{"s3://foo/bar/baz", "s3://foo/bar/baz"},
		{"s3://foo/bar/baz?type=hello/world", "s3://foo/bar/baz"},
		{"s3://foo/bar/baz?region=us-east-1", "s3://foo/bar/baz?region=us-east-1"},
		{"s3://foo/bar/baz?disableSSL=true&type=text/csv", "s3://foo/bar/baz?disableSSL=true"},
		{"s3://foo/bar/baz?type=text/csv&s3ForcePathStyle=true&endpoint=1.2.3.4", "s3://foo/bar/baz?endpoint=1.2.3.4&s3ForcePathStyle=true"},
		{"gs://foo/bar/baz", "gs://foo/bar/baz"},
		{"gs://foo/bar/baz?type=foo/bar", "gs://foo/bar/baz"},
		{"gs://foo/bar/baz?access_id=123", "gs://foo/bar/baz?access_id=123"},
		{"gs://foo/bar/baz?private_key_path=/foo/bar", "gs://foo/bar/baz?private_key_path=%2Ffoo%2Fbar"},
		{"gs://foo/bar/baz?private_key_path=key.json&foo=bar", "gs://foo/bar/baz?private_key_path=key.json"},
		{"gs://foo/bar/baz?private_key_path=key.json&foo=bar&access_id=abcd", "gs://foo/bar/baz?access_id=abcd&private_key_path=key.json"},
	}

	for _, d := range data {
		u, _ := url.Parse(d.in)
		out := sanitizeURL(u).String()
		assert.Equal(t, d.expected, out)
	}
}
