package data

import (
	"context"
	"io/fs"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/hairyhenderson/gomplate/v4/internal/datafs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadMerge(t *testing.T) {
	ctx := context.Background()

	jsonContent := `{"hello": "world"}`
	yamlContent := "hello: earth\ngoodnight: moon\n"
	arrayContent := `["hello", "world"]`

	mergedContent := "goodnight: moon\nhello: world\n"

	wd, _ := os.Getwd()

	// MapFS doesn't support windows path separators, so we use / exclusively
	// in this test
	vol := filepath.VolumeName(wd)
	if vol != "" && wd != vol {
		wd = wd[len(vol)+1:]
	} else if wd[0] == '/' {
		wd = wd[1:]
	}
	wd = filepath.ToSlash(wd)

	fsys := datafs.WrapWdFS(fstest.MapFS{
		"tmp":                          {Mode: fs.ModeDir | 0o777},
		"tmp/jsonfile.json":            {Data: []byte(jsonContent)},
		"tmp/array.json":               {Data: []byte(arrayContent)},
		"tmp/yamlfile.yaml":            {Data: []byte(yamlContent)},
		"tmp/textfile.txt":             {Data: []byte(`plain text...`)},
		path.Join(wd, "jsonfile.json"): {Data: []byte(jsonContent)},
		path.Join(wd, "array.json"):    {Data: []byte(arrayContent)},
		path.Join(wd, "yamlfile.yaml"): {Data: []byte(yamlContent)},
		path.Join(wd, "textfile.txt"):  {Data: []byte(`plain text...`)},
	})

	source := &Source{Alias: "foo", URL: mustParseURL("merge:file:///tmp/jsonfile.json|file:///tmp/yamlfile.yaml")}
	source.fs = fsys
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

	source.URL = mustParseURL("merge:jsonfile.json|baz")
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
