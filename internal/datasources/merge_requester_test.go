package datasources

import (
	"context"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/hairyhenderson/gomplate/v3/internal/config"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestReadMerge(t *testing.T) {
	registerRequesters()

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

	u := mustParseURL("merge:file:///tmp/jsonfile.json|file:///tmp/yamlfile.yaml")

	r := &mergeRequester{
		ds: map[string]config.DataSource{
			"foo":       {URL: u},
			"bar":       {URL: mustParseURL("file:///tmp/jsonfile.json")},
			"baz":       {URL: mustParseURL("file:///tmp/yamlfile.yaml")},
			"text":      {URL: mustParseURL("file:///tmp/textfile.txt")},
			"badscheme": {URL: mustParseURL("bad:///scheme.json")},
			"badtype":   {URL: mustParseURL("file:///tmp/textfile.txt?type=foo/bar")},
			"array":     {URL: mustParseURL("file:///tmp/array.json?type=" + url.QueryEscape(jsonArrayMimetype))},
		},
	}

	actual, err := r.Request(ctx, u, nil)
	assert.NoError(t, err)
	b, _ := ioutil.ReadAll(actual.Body)
	assert.Equal(t, mergedContent, string(b))
	assert.Equal(t, yamlMimetype, actual.ContentType)

	u = mustParseURL("merge:bar|baz")
	actual, err = r.Request(ctx, u, nil)
	assert.NoError(t, err)
	b, _ = ioutil.ReadAll(actual.Body)
	assert.Equal(t, mergedContent, string(b))
	assert.Equal(t, yamlMimetype, actual.ContentType)

	u = mustParseURL("merge:./jsonfile.json|baz")
	actual, err = r.Request(ctx, u, nil)
	assert.NoError(t, err)
	b, _ = ioutil.ReadAll(actual.Body)
	assert.Equal(t, mergedContent, string(b))
	assert.Equal(t, yamlMimetype, actual.ContentType)

	oldCtx := ctx
	ctx = config.WithDataSources(ctx, map[string]config.DataSource{
		"incontext": {URL: mustParseURL("file:///tmp/jsonfile.json")},
	})
	u = mustParseURL("merge:incontext|baz")
	actual, err = r.Request(ctx, u, nil)
	assert.NoError(t, err)
	b, _ = ioutil.ReadAll(actual.Body)
	assert.Equal(t, mergedContent, string(b))
	assert.Equal(t, yamlMimetype, actual.ContentType)

	ctx = oldCtx
	// it must be added by now
	u = mustParseURL("merge:incontext|baz")
	actual, err = r.Request(ctx, u, nil)
	assert.NoError(t, err)
	b, _ = ioutil.ReadAll(actual.Body)
	assert.Equal(t, mergedContent, string(b))
	assert.Equal(t, yamlMimetype, actual.ContentType)

	u = mustParseURL("merge:file:///tmp/jsonfile.json")
	_, err = r.Request(ctx, u, nil)
	assert.Error(t, err)

	u = mustParseURL("merge:bogusalias|file:///tmp/jsonfile.json")
	_, err = r.Request(ctx, u, nil)
	assert.Error(t, err)

	u = mustParseURL("merge:file:///tmp/jsonfile.json|badscheme")
	_, err = r.Request(ctx, u, nil)
	assert.Error(t, err)

	u = mustParseURL("merge:file:///tmp/jsonfile.json|badtype")
	_, err = r.Request(ctx, u, nil)
	assert.Error(t, err)

	u = mustParseURL("merge:file:///tmp/jsonfile.json|array")
	_, err = r.Request(ctx, u, nil)
	assert.Error(t, err)
}

func TestMergeData(t *testing.T) {
	def := map[string]interface{}{
		"f": true,
		"t": false,
		"z": "def",
	}
	out, err := mergeData([]map[string]interface{}{def})
	assert.NoError(t, err)
	assert.Equal(t, "f: true\nt: false\nz: def\n", string(out))

	over := map[string]interface{}{
		"f": false,
		"t": true,
		"z": "over",
	}
	out, err = mergeData([]map[string]interface{}{over, def})
	assert.NoError(t, err)
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
	assert.NoError(t, err)
	assert.Equal(t, "f: false\nm:\n  a: aaa\nt: true\nz: over\n", string(out))

	uber := map[string]interface{}{
		"z": "über",
	}
	out, err = mergeData([]map[string]interface{}{uber, over, def})
	assert.NoError(t, err)
	assert.Equal(t, "f: false\nm:\n  a: aaa\nt: true\nz: über\n", string(out))

	uber = map[string]interface{}{
		"m": "notamap",
		"z": map[string]interface{}{
			"b": "bbb",
		},
	}
	out, err = mergeData([]map[string]interface{}{uber, over, def})
	assert.NoError(t, err)
	assert.Equal(t, "f: false\nm: notamap\nt: true\nz:\n  b: bbb\n", string(out))

	uber = map[string]interface{}{
		"m": map[string]interface{}{
			"b": "bbb",
		},
	}
	out, err = mergeData([]map[string]interface{}{uber, over, def})
	assert.NoError(t, err)
	assert.Equal(t, "f: false\nm:\n  a: aaa\n  b: bbb\nt: true\nz: over\n", string(out))
}
