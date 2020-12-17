package data

import (
	"context"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/hairyhenderson/gomplate/v3/internal/config"
	"github.com/hairyhenderson/gomplate/v3/internal/datasources"
	"github.com/spf13/afero"

	"github.com/stretchr/testify/assert"
)

func TestReadMerge(t *testing.T) {
	jsonContent := `{"hello": "world"}`
	yamlContent := "hello: earth\ngoodnight: moon\n"
	arrayContent := `["hello", "world"]`

	mergedContent := "goodnight: moon\nhello: world\n"

	fs := afero.NewMemMapFs()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = config.WithFileSystem(ctx, fs)

	_ = fs.Mkdir("/tmp", 0777)
	f, _ := fs.Create("/tmp/jsonfile.json")
	_, _ = f.WriteString(jsonContent)
	f, _ = fs.Create("/tmp/array.json")
	_, _ = f.WriteString(arrayContent)
	f, _ = fs.Create("/tmp/yamlfile.yaml")
	_, _ = f.WriteString(yamlContent)
	f, _ = fs.Create("/tmp/textfile.txt")
	_, _ = f.WriteString(`plain text...`)

	wd, _ := os.Getwd()
	_ = fs.Mkdir(wd, 0777)
	f, _ = fs.Create(filepath.Join(wd, "jsonfile.json"))
	_, _ = f.WriteString(jsonContent)
	f, _ = fs.Create(filepath.Join(wd, "array.json"))
	_, _ = f.WriteString(arrayContent)
	f, _ = fs.Create(filepath.Join(wd, "yamlfile.yaml"))
	_, _ = f.WriteString(yamlContent)
	f, _ = fs.Create(filepath.Join(wd, "textfile.txt"))
	_, _ = f.WriteString(`plain text...`)

	source := config.DataSource{URL: mustParseURL("merge:file:///tmp/jsonfile.json|file:///tmp/yamlfile.yaml")}

	ctx = config.WithDataSources(ctx, map[string]config.DataSource{
		"foo":       source,
		"bar":       {URL: mustParseURL("file:///tmp/jsonfile.json")},
		"baz":       {URL: mustParseURL("file:///tmp/yamlfile.yaml")},
		"text":      {URL: mustParseURL("file:///tmp/textfile.txt")},
		"badscheme": {URL: mustParseURL("bad:///scheme.json")},
		"badtype":   {URL: mustParseURL("file:///tmp/textfile.txt?type=foo/bar")},
		"array":     {URL: mustParseURL("file:///tmp/array.json?type=" + url.QueryEscape(jsonArrayMimetype))},
	})

	mt, actual, err := datasources.ReadDataSource(ctx, source)
	assert.NoError(t, err)
	assert.Equal(t, mergedContent, string(actual))
	assert.Equal(t, yamlMimetype, mt)

	source.URL = mustParseURL("merge:bar|baz")
	mt, actual, err = datasources.ReadDataSource(ctx, source)
	assert.NoError(t, err)
	assert.Equal(t, mergedContent, string(actual))
	assert.Equal(t, yamlMimetype, mt)

	source.URL = mustParseURL("merge:./jsonfile.json|baz")
	mt, actual, err = datasources.ReadDataSource(ctx, source)
	assert.NoError(t, err)
	assert.Equal(t, mergedContent, string(actual))
	assert.Equal(t, yamlMimetype, mt)

	source.URL = mustParseURL("merge:file:///tmp/jsonfile.json")
	_, _, err = datasources.ReadDataSource(ctx, source)
	assert.Error(t, err)

	source.URL = mustParseURL("merge:bogusalias|file:///tmp/jsonfile.json")
	_, _, err = datasources.ReadDataSource(ctx, source)
	assert.Error(t, err)

	source.URL = mustParseURL("merge:file:///tmp/jsonfile.json|badscheme")
	_, _, err = datasources.ReadDataSource(ctx, source)
	assert.Error(t, err)

	source.URL = mustParseURL("merge:file:///tmp/jsonfile.json|badtype")
	_, _, err = datasources.ReadDataSource(ctx, source)
	assert.Error(t, err)

	source.URL = mustParseURL("merge:file:///tmp/jsonfile.json|array")
	_, _, err = datasources.ReadDataSource(ctx, source)
	assert.Error(t, err)
}
