package data

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/hairyhenderson/gomplate/v3/internal/config"
	"github.com/hairyhenderson/gomplate/v3/internal/datasource"
	"github.com/pkg/errors"

	"github.com/stretchr/testify/assert"
)

const osWindows = "windows"

type mockSource struct {
	err    error
	data   *datasource.Data
	header http.Header

	args []string
}

func (m *mockSource) Read(ctx context.Context, args ...string) (*datasource.Data, error) {
	m.args = args
	return m.data, m.err
}

func (m *mockSource) Cleanup() {
}

type mockSourceReg map[string]datasource.Source

var _ datasource.SourceRegistry = (*mockSourceReg)(nil)

func (m mockSourceReg) Register(alias string, url *url.URL, header http.Header) (datasource.Source, error) {
	s := &mockSource{
		data: &datasource.Data{
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
func (m mockSourceReg) Get(alias string) datasource.Source {
	return m[alias]
}

// Dynamic registers a new dynamically-defined source - the alias would be a URL in this case
func (m mockSourceReg) Dynamic(alias string, header http.Header) (datasource.Source, error) {
	u, err := url.Parse(alias)
	if err != nil || !u.IsAbs() {
		return nil, fmt.Errorf("invalid: %w", err)
	}
	return m.Register(alias, u, header)
}

func TestDatasource(t *testing.T) {
	setup := func(ext, mime string, contents []byte) *Data {
		fname := "foo." + ext
		var uPath string
		if runtime.GOOS == osWindows {
			uPath = "C:/tmp/" + fname
		} else {
			uPath = "/tmp/" + fname
		}

		return &Data{sourceReg: mockSourceReg{
			"foo": &mockSource{
				data: &datasource.Data{
					Bytes: contents,
					URL:   &url.URL{Scheme: "file", Path: uPath},
					MType: mime,
				},
			},
		}}
	}

	test := func(ext, mime string, contents []byte) {
		d := setup(ext, mime, contents)
		expected := map[string]interface{}{"hello": map[string]interface{}{"cruel": "world"}}
		actual, err := d.Datasource("foo")
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	}

	test("json", jsonMimetype, []byte(`{"hello":{"cruel":"world"}}`))
	test("yml", yamlMimetype, []byte("hello:\n  cruel: world\n"))

	d := setup("", textMimetype, nil)
	actual, err := d.Datasource("foo")
	assert.NoError(t, err)
	assert.Equal(t, "", actual)

	_, err = d.Datasource("bar")
	assert.Error(t, err)
}

func TestDatasourceReachable(t *testing.T) {
	data := &Data{sourceReg: mockSourceReg{
		"foo": &mockSource{},
		"bar": &mockSource{err: errors.New("foo")},
	}}

	assert.True(t, data.DatasourceReachable("foo"))
	assert.False(t, data.DatasourceReachable("bar"))
	assert.False(t, data.DatasourceReachable("baz"))
}

func TestDatasourceExists(t *testing.T) {
	data := &Data{sourceReg: mockSourceReg{
		"foo": &mockSource{},
	}}
	assert.True(t, data.DatasourceExists("foo"))
	assert.False(t, data.DatasourceExists("bar"))
}

func TestInclude(t *testing.T) {
	contents := "hello world"
	data := &Data{
		sourceReg: mockSourceReg{
			"foo": &mockSource{
				data: &datasource.Data{Bytes: []byte(contents)},
			},
		},
	}
	actual, err := data.Include("foo")
	assert.NoError(t, err)
	assert.Equal(t, contents, actual)
}

type errorReader struct{}

func (e errorReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("error")
}

// nolint: megacheck
func TestDefineDatasource(t *testing.T) {
	d := &Data{sourceReg: mockSourceReg{}}
	_, err := d.DefineDatasource("", "foo.json")
	assert.Error(t, err)

	d = &Data{sourceReg: mockSourceReg{}}
	_, err = d.DefineDatasource("", "../foo.json")
	assert.Error(t, err)

	d = &Data{sourceReg: mockSourceReg{}}
	_, err = d.DefineDatasource("", "ftp://example.com/foo.yml")
	assert.Error(t, err)

	d = &Data{sourceReg: mockSourceReg{}}
	_, err = d.DefineDatasource("data", "foo.json")
	ok := d.sourceReg.Exists("data")
	assert.NoError(t, err)
	assert.True(t, ok)

	d = &Data{sourceReg: mockSourceReg{}}
	_, err = d.DefineDatasource("data", "/otherdir/foo.json")
	ok = d.sourceReg.Exists("data")
	assert.NoError(t, err)
	assert.True(t, ok)

	d = &Data{sourceReg: mockSourceReg{}}
	_, err = d.DefineDatasource("data", "sftp://example.com/blahblah/foo.json")
	ok = d.sourceReg.Exists("data")
	assert.NoError(t, err)
	assert.True(t, ok)

	d = &Data{
		sourceReg: mockSourceReg{
			"data": &mockSource{},
		},
	}
	_, err = d.DefineDatasource("data", "/otherdir/foo.json")
	ok = d.sourceReg.Exists("data")
	assert.NoError(t, err)
	assert.True(t, ok)

	d = &Data{sourceReg: mockSourceReg{}}
	_, err = d.DefineDatasource("data", "/otherdir/foo?type=application/x-env")
	s := d.sourceReg.Get("data")
	assert.NoError(t, err)
	assert.True(t, ok)
	dd, err := s.Read(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, dd)
	m, err := dd.MediaType()
	assert.Equal(t, "application/x-env", m)
}

func TestMimeType(t *testing.T) {
	s := &Source{URL: mustParseURL("http://example.com/list?type=a/b/c")}
	_, err := s.mimeType("")
	assert.Error(t, err)

	data := []struct {
		url       string
		mediaType string
		expected  string
	}{
		{"http://example.com/foo.json",
			"",
			jsonMimetype},
		{"http://example.com/foo.json",
			"text/foo",
			"text/foo"},
		{"http://example.com/foo.json?type=application/yaml",
			"text/foo",
			"application/yaml"},
		{"http://example.com/list?type=application/array%2Bjson",
			"text/foo",
			"application/array+json"},
		{"http://example.com/list?type=application/array+json",
			"",
			"application/array+json"},
		{"http://example.com/unknown",
			"",
			"text/plain"},
	}

	for i, d := range data {
		t.Run(fmt.Sprintf("%d:%q,%q==%q", i, d.url, d.mediaType, d.expected), func(t *testing.T) {
			s := &Source{URL: mustParseURL(d.url), mediaType: d.mediaType}
			mt, err := s.mimeType("")
			assert.NoError(t, err)
			assert.Equal(t, d.expected, mt)
		})
	}
}

func TestMimeTypeWithArg(t *testing.T) {
	s := &Source{URL: mustParseURL("http://example.com")}
	_, err := s.mimeType("h\nttp://foo")
	assert.Error(t, err)

	data := []struct {
		url       string
		mediaType string
		arg       string
		expected  string
	}{
		{"http://example.com/unknown",
			"",
			"/foo.json",
			"application/json"},
		{"http://example.com/unknown",
			"",
			"foo.json",
			"application/json"},
		{"http://example.com/",
			"text/foo",
			"/foo.json",
			"text/foo"},
		{"git+https://example.com/myrepo",
			"",
			"//foo.yaml",
			"application/yaml"},
		{"http://example.com/foo.json",
			"",
			"/foo.yaml",
			"application/yaml"},
		{"http://example.com/foo.json?type=application/array+yaml",
			"",
			"/foo.yaml",
			"application/array+yaml"},
		{"http://example.com/foo.json?type=application/array+yaml",
			"",
			"/foo.yaml?type=application/yaml",
			"application/yaml"},
		{"http://example.com/foo.json?type=application/array+yaml",
			"text/plain",
			"/foo.yaml?type=application/yaml",
			"application/yaml"},
	}

	for i, d := range data {
		t.Run(fmt.Sprintf("%d:%q,%q,%q==%q", i, d.url, d.mediaType, d.arg, d.expected), func(t *testing.T) {
			s := &Source{URL: mustParseURL(d.url), mediaType: d.mediaType}
			mt, err := s.mimeType(d.arg)
			assert.NoError(t, err)
			assert.Equal(t, d.expected, mt)
		})
	}
}

func TestQueryParse(t *testing.T) {
	expected := &url.URL{
		Scheme:   "http",
		Host:     "example.com",
		Path:     "/foo.json",
		RawQuery: "bar",
	}
	u, err := parseSourceURL("http://example.com/foo.json?bar")
	assert.NoError(t, err)
	assert.EqualValues(t, expected, u)
}

func TestAbsFileURL(t *testing.T) {
	cwd, _ := os.Getwd()
	// make this pass on Windows
	cwd = filepath.ToSlash(cwd)
	expected := &url.URL{
		Scheme: "file",
		Host:   "",
		Path:   "/tmp/foo",
	}
	u, err := absFileURL("/tmp/foo")
	assert.NoError(t, err)
	assert.EqualValues(t, expected, u)

	expected = &url.URL{
		Scheme: "file",
		Host:   "",
		Path:   cwd + "/tmp/foo",
	}
	u, err = absFileURL("tmp/foo")
	assert.NoError(t, err)
	assert.EqualValues(t, expected, u)

	expected = &url.URL{
		Scheme:   "file",
		Host:     "",
		Path:     cwd + "/tmp/foo",
		RawQuery: "q=p",
	}
	u, err = absFileURL("tmp/foo?q=p")
	assert.NoError(t, err)
	assert.EqualValues(t, expected, u)
}

func TestFromConfig(t *testing.T) {
	defer func() { sourceRegistry = datasource.DefaultRegistry }()
	sreg := mockSourceReg{}
	sourceRegistry = sreg

	cfg := &config.Config{}

	_, err := FromConfig(cfg)
	assert.NoError(t, err)
	assert.Empty(t, sreg)
	// assert.EqualValues(t, expected, d)

	cfg = &config.Config{
		DataSources: map[string]config.DataSource{
			"foo": {
				URL: mustParseURL("http://example.com"),
			},
		},
	}

	sreg = mockSourceReg{}
	sourceRegistry = sreg

	_, err = FromConfig(cfg)
	assert.NoError(t, err)
	assert.EqualValues(t, &mockSource{
		data: &datasource.Data{
			URL: mustParseURL("http://example.com"),
		},
	}, sreg["foo"])

	cfg = &config.Config{
		DataSources: map[string]config.DataSource{
			"foo": {
				URL: mustParseURL("http://foo.com"),
			},
		},
		Context: map[string]config.DataSource{
			"bar": {
				URL: mustParseURL("http://bar.com"),
				Header: http.Header{
					"Foo": []string{"bar"},
				},
			},
		},
		ExtraHeaders: map[string]http.Header{
			"baz": {
				"Foo": []string{"bar"},
			},
		},
	}
	sreg = mockSourceReg{}
	sourceRegistry = sreg

	expected := &Data{
		sourceReg: sreg,
		extraHeaders: map[string]http.Header{
			"baz": {
				"Foo": []string{"bar"},
			},
		},
	}
	d, err := FromConfig(cfg)
	assert.NoError(t, err)
	assert.EqualValues(t, expected, d)
	assert.EqualValues(t, mockSourceReg{
		"foo": &mockSource{
			data: &datasource.Data{
				URL: mustParseURL("http://foo.com"),
			},
		},
		"bar": &mockSource{
			data: &datasource.Data{
				URL: mustParseURL("http://bar.com"),
			},
			header: http.Header{
				"Foo": []string{"bar"},
			},
		}}, sreg)
}

func mustParseURL(in string) *url.URL {
	u, _ := url.Parse(in)
	return u
}
