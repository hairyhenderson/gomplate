package data

import (
	"context"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadMerge(t *testing.T) {
	ctx := context.Background()

	jsonContent := `{"hello": "world"}`
	yamlContent := "hello: earth\ngoodnight: moon\n"
	arrayContent := `["hello", "world"]`

	mergedContent := "goodnight: moon\nhello: world\n"

	fs := afero.NewMemMapFs()

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

	actual, err := d.readMerge(ctx, source)
	require.NoError(t, err)
	assert.Equal(t, mergedContent, string(actual))

	source.URL = mustParseURL("merge:bar|baz")
	actual, err = d.readMerge(ctx, source)
	require.NoError(t, err)
	assert.Equal(t, mergedContent, string(actual))

	source.URL = mustParseURL("merge:./jsonfile.json|baz")
	actual, err = d.readMerge(ctx, source)
	require.NoError(t, err)
	assert.Equal(t, mergedContent, string(actual))

	source.URL = mustParseURL("merge:file:///tmp/jsonfile.json")
	_, err = d.readMerge(ctx, source)
	assert.Error(t, err)

	source.URL = mustParseURL("merge:bogusalias|file:///tmp/jsonfile.json")
	_, err = d.readMerge(ctx, source)
	assert.Error(t, err)

	source.URL = mustParseURL("merge:file:///tmp/jsonfile.json|badscheme")
	_, err = d.readMerge(ctx, source)
	assert.Error(t, err)

	source.URL = mustParseURL("merge:file:///tmp/jsonfile.json|badtype")
	_, err = d.readMerge(ctx, source)
	assert.Error(t, err)

	source.URL = mustParseURL("merge:file:///tmp/jsonfile.json|array")
	_, err = d.readMerge(ctx, source)
	assert.Error(t, err)
}

func TestMergeData(t *testing.T) {
	def := map[string]interface{}{
		"f": true,
		"t": false,
		"z": "def",
	}
	out, err := mergeData([]map[string]interface{}{def})
	require.NoError(t, err)
	assert.Equal(t, "f: true\nt: false\nz: def\n", string(out))

	over := map[string]interface{}{
		"f": false,
		"t": true,
		"z": "over",
	}
	out, err = mergeData([]map[string]interface{}{over, def})
	require.NoError(t, err)
	assert.Equal(t, "f: false\nt: true\nz: over\n", string(out))

	over = map[string]interface{}{
		"f": false,
		"t": true,
		"z": "over",
		"m": map[string]interface{}{
			"a": "aaa",
		},
	}
	out, err = mergeData([]map[string]interface{}{over, def})
	require.NoError(t, err)
	assert.Equal(t, "f: false\nm:\n  a: aaa\nt: true\nz: over\n", string(out))

	uber := map[string]interface{}{
		"z": "über",
	}
	out, err = mergeData([]map[string]interface{}{uber, over, def})
	require.NoError(t, err)
	assert.Equal(t, "f: false\nm:\n  a: aaa\nt: true\nz: über\n", string(out))

	uber = map[string]interface{}{
		"m": "notamap",
		"z": map[string]interface{}{
			"b": "bbb",
		},
	}
	out, err = mergeData([]map[string]interface{}{uber, over, def})
	require.NoError(t, err)
	assert.Equal(t, "f: false\nm: notamap\nt: true\nz:\n  b: bbb\n", string(out))

	uber = map[string]interface{}{
		"m": map[string]interface{}{
			"b": "bbb",
		},
	}
	out, err = mergeData([]map[string]interface{}{uber, over, def})
	require.NoError(t, err)
	assert.Equal(t, "f: false\nm:\n  a: aaa\n  b: bbb\nt: true\nz: over\n", string(out))
}
