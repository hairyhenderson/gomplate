package datasource

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"

	"github.com/stretchr/testify/assert"
)

type mockSource struct {
	err    error
	data   *Data
	header http.Header

	args []string
}

func (m *mockSource) Read(ctx context.Context, args ...string) (*Data, error) {
	m.args = args
	return m.data, m.err
}

func (m *mockSource) Cleanup() {
}

type mockSourceReg map[string]Source

var _ SourceRegistry = (*mockSourceReg)(nil)

func (m mockSourceReg) Register(alias string, url *url.URL, header http.Header) (Source, error) {
	s := &mockSource{
		data: &Data{
			URL: url,
		},
		header: header,
	}
	m[alias] = s
	return s, nil
}

func (m mockSourceReg) Exists(alias string) bool {
	_, ok := m[alias]
	return ok
}

// Get returns a cached source if it exists
func (m mockSourceReg) Get(alias string) Source {
	return m[alias]
}

// Dynamic registers a new dynamically-defined source - the alias would be a URL in this case
func (m mockSourceReg) Dynamic(alias string, header http.Header) (Source, error) {
	u, err := url.Parse(alias)
	if err != nil || !u.IsAbs() {
		return nil, fmt.Errorf("invalid: %w", err)
	}
	return m.Register(alias, u, header)
}

func TestReadMerge(t *testing.T) {
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

	fileReader := &File{fs}

	readers = map[string]Reader{
		"file": fileReader,
	}
	defer func() { readers = nil }()

	source := &src{
		alias: "foo",
		url:   mustParseURL("merge:file:///tmp/jsonfile.json|file:///tmp/yamlfile.yaml"),
		r:     fileReader,
	}

	ctx := context.Background()
	m := &Merge{reg: &srcRegistry{
		sources: map[string]Source{
			"foo": source,
			"bar": &src{
				alias: "bar",
				url:   mustParseURL("file:///tmp/jsonfile.json"),
				r:     fileReader,
			},
			"baz": &src{
				alias: "baz",
				url:   mustParseURL("file:///tmp/yamlfile.yaml"),
				r:     fileReader,
			},
			"text": &src{
				alias: "text",
				url:   mustParseURL("file:///tmp/textfile.txt"),
				r:     fileReader,
			},
			"badscheme": &src{
				alias: "badscheme",
				url:   mustParseURL("bad:///scheme.json"),
				r:     fileReader,
			},
			"badtype": &src{
				alias: "badtype",
				url:   mustParseURL("file:///tmp/textfile.txt?type=foo/bar"),
				r:     fileReader,
			},
			"array": &src{
				alias: "array",
				url:   mustParseURL("file:///tmp/array.json?type=" + url.QueryEscape(jsonArrayMimetype)),
				r:     fileReader,
			},
		},
	}}

	actual, err := m.Read(ctx, mustParseURL("merge:file:///tmp/jsonfile.json|file:///tmp/yamlfile.yaml"))
	assert.NoError(t, err)
	assert.Equal(t, mergedContent, string(actual.Bytes))

	actual, err = m.Read(ctx, mustParseURL("merge:bar|baz"))
	assert.NoError(t, err)
	assert.Equal(t, mergedContent, string(actual.Bytes))

	actual, err = m.Read(ctx, mustParseURL("merge:./jsonfile.json|baz"))
	assert.NoError(t, err)
	assert.Equal(t, mergedContent, string(actual.Bytes))

	_, err = m.Read(ctx, mustParseURL("merge:file:///tmp/jsonfile.json"))
	assert.Error(t, err)

	_, err = m.Read(ctx, mustParseURL("merge:bogusalias|file:///tmp/jsonfile.json"))
	assert.Error(t, err)

	_, err = m.Read(ctx, mustParseURL("merge:file:///tmp/jsonfile.json|badscheme"))
	assert.Error(t, err)

	_, err = m.Read(ctx, mustParseURL("merge:file:///tmp/jsonfile.json|badtype"))
	assert.Error(t, err)

	_, err = m.Read(ctx, mustParseURL("merge:file:///tmp/jsonfile.json|array"))
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
