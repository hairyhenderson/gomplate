//+build !windows

package data

import (
	"net/url"
	"testing"

	"github.com/blang/vfs"
	"github.com/blang/vfs/memfs"
	"github.com/stretchr/testify/assert"
)

func TestReadMerge(t *testing.T) {
	jsonContent := []byte(`{"hello": "world"}`)
	yamlContent := []byte("hello: earth\ngoodnight: moon\n")
	arrayContent := []byte(`["hello", "world"]`)

	mergedContent := []byte("goodnight: moon\nhello: world\n")

	fs := memfs.Create()

	_ = fs.Mkdir("/tmp", 0777)
	f, _ := vfs.Create(fs, "/tmp/jsonfile.json")
	_, _ = f.Write(jsonContent)
	f, _ = vfs.Create(fs, "/tmp/array.json")
	_, _ = f.Write(arrayContent)
	f, _ = vfs.Create(fs, "/tmp/yamlfile.yaml")
	_, _ = f.Write(yamlContent)
	f, _ = vfs.Create(fs, "/tmp/textfile.txt")
	_, _ = f.Write([]byte(`plain text...`))

	source := &Source{Alias: "foo", URL: mustParseURL("merge:file:///tmp/jsonfile.json|file:///tmp/yamlfile.yaml")}
	source.fs = fs
	d := &Data{
		Sources: map[string]*Source{
			"foo":       source,
			"bar":       {Alias: "bar", URL: mustParseURL("file:///tmp/jsonfile.json")},
			"baz":       {Alias: "baz", URL: mustParseURL("file:///tmp/yamlfile.yaml")},
			"text":      {Alias: "text", URL: mustParseURL("file:///tmp/textfile.txt")},
			"badscheme": {Alias: "badscheme", URL: mustParseURL("bad:///scheme.json")},
			"badtype":   {Alias: "badtype", URL: mustParseURL("file:///tmp/textfile.txt?type=foo/bar")},
			"array":     {Alias: "array", URL: mustParseURL("file:///tmp/array.json?type=" + url.QueryEscape(jsonArrayMimetype))},
		},
	}

	actual, err := d.readMerge(source)
	assert.NoError(t, err)
	assert.Equal(t, mergedContent, actual)

	source.URL = mustParseURL("merge:bar|baz")
	actual, err = d.readMerge(source)
	assert.NoError(t, err)
	assert.Equal(t, mergedContent, actual)

	source.URL = mustParseURL("merge:file:///tmp/jsonfile.json")
	_, err = d.readMerge(source)
	assert.Error(t, err)

	source.URL = mustParseURL("merge:bogusalias|file:///tmp/jsonfile.json")
	_, err = d.readMerge(source)
	assert.Error(t, err)

	source.URL = mustParseURL("merge:file:///tmp/jsonfile.json|badscheme")
	_, err = d.readMerge(source)
	assert.Error(t, err)

	source.URL = mustParseURL("merge:file:///tmp/jsonfile.json|badtype")
	_, err = d.readMerge(source)
	assert.Error(t, err)

	source.URL = mustParseURL("merge:file:///tmp/jsonfile.json|array")
	_, err = d.readMerge(source)
	assert.Error(t, err)
}
