// +build !windows

package data

import (
	"net/url"
	"testing"

	"github.com/blang/vfs"
	"github.com/blang/vfs/memfs"
	"github.com/stretchr/testify/assert"
)

func mustParseURL(in string) *url.URL {
	u, _ := url.Parse(in)
	return u
}

func TestReadFile(t *testing.T) {
	content := []byte(`hello world`)
	fs := memfs.Create()

	_ = fs.Mkdir("/tmp", 0777)
	f, _ := vfs.Create(fs, "/tmp/foo")
	_, _ = f.Write(content)

	_ = fs.Mkdir("/tmp/partial", 0777)
	f, _ = vfs.Create(fs, "/tmp/partial/foo.txt")
	_, _ = f.Write(content)
	_, _ = vfs.Create(fs, "/tmp/partial/bar.txt")
	_, _ = vfs.Create(fs, "/tmp/partial/baz.txt")

	source, _ := NewSource("foo", mustParseURL("file:///tmp/foo"))
	source.FS = fs

	actual, err := readFile(source)
	assert.NoError(t, err)
	assert.Equal(t, content, actual)

	source, _ = NewSource("bogus", mustParseURL("file:///bogus"))
	source.FS = fs
	_, err = readFile(source)
	assert.Error(t, err)

	source, _ = NewSource("partial", mustParseURL("file:///tmp/partial"))
	source.FS = fs
	actual, err = readFile(source, "foo.txt")
	assert.NoError(t, err)
	assert.Equal(t, content, actual)

	source, _ = NewSource("dir", mustParseURL("file:///tmp/partial/"))
	source.FS = fs
	actual, err = readFile(source)
	assert.NoError(t, err)
	assert.Equal(t, []byte(`["bar.txt","baz.txt","foo.txt"]`), actual)
}
