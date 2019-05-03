package data

import (
	"testing"

	"github.com/spf13/afero"

	"github.com/stretchr/testify/assert"
)

func TestReadFile(t *testing.T) {
	content := []byte(`hello world`)
	fs := afero.NewMemMapFs()

	_ = fs.Mkdir("/tmp", 0777)
	f, _ := fs.Create("/tmp/foo")
	_, _ = f.Write(content)

	_ = fs.Mkdir("/tmp/partial", 0777)
	f, _ = fs.Create("/tmp/partial/foo.txt")
	_, _ = f.Write(content)
	_, _ = fs.Create("/tmp/partial/bar.txt")
	_, _ = fs.Create("/tmp/partial/baz.txt")

	source := &Source{Alias: "foo", URL: mustParseURL("file:///tmp/foo")}
	source.fs = fs

	actual, err := readFile(source)
	assert.NoError(t, err)
	assert.Equal(t, content, actual)

	source = &Source{Alias: "bogus", URL: mustParseURL("file:///bogus")}
	source.fs = fs
	_, err = readFile(source)
	assert.Error(t, err)

	source = &Source{Alias: "partial", URL: mustParseURL("file:///tmp/partial")}
	source.fs = fs
	actual, err = readFile(source, "foo.txt")
	assert.NoError(t, err)
	assert.Equal(t, content, actual)

	source = &Source{Alias: "dir", URL: mustParseURL("file:///tmp/partial/")}
	source.fs = fs
	actual, err = readFile(source)
	assert.NoError(t, err)
	assert.Equal(t, []byte(`["bar.txt","baz.txt","foo.txt"]`), actual)

	source = &Source{Alias: "dir", URL: mustParseURL("file:///tmp/partial/?type=application/json")}
	source.fs = fs
	actual, err = readFile(source)
	assert.NoError(t, err)
	assert.Equal(t, []byte(`["bar.txt","baz.txt","foo.txt"]`), actual)
	mime, err := source.mimeType()
	assert.NoError(t, err)
	assert.Equal(t, "application/json", mime)

	source = &Source{Alias: "dir", URL: mustParseURL("file:///tmp/partial/?type=application/json")}
	source.fs = fs
	actual, err = readFile(source, "foo.txt")
	assert.NoError(t, err)
	assert.Equal(t, content, actual)
	mime, err = source.mimeType()
	assert.NoError(t, err)
	assert.Equal(t, "application/json", mime)
}
