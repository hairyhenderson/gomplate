package data

import (
	"fmt"
	"net/url"
	"runtime"
	"strings"
	"testing"

	"github.com/spf13/afero"

	"github.com/stretchr/testify/assert"
)

const osWindows = "windows"

func TestNewData(t *testing.T) {
	d, err := NewData(nil, nil)
	assert.NoError(t, err)
	assert.Len(t, d.Sources, 0)

	d, err = NewData([]string{"foo=http:///foo.json"}, nil)
	assert.NoError(t, err)
	assert.Equal(t, "/foo.json", d.Sources["foo"].URL.Path)

	d, err = NewData([]string{"foo=http:///foo.json"}, []string{})
	assert.NoError(t, err)
	assert.Equal(t, "/foo.json", d.Sources["foo"].URL.Path)
	assert.Empty(t, d.Sources["foo"].header)

	d, err = NewData([]string{"foo=http:///foo.json"}, []string{"bar=Accept: blah"})
	assert.NoError(t, err)
	assert.Equal(t, "/foo.json", d.Sources["foo"].URL.Path)
	assert.Empty(t, d.Sources["foo"].header)

	d, err = NewData([]string{"foo=http:///foo.json"}, []string{"foo=Accept: blah"})
	assert.NoError(t, err)
	assert.Equal(t, "/foo.json", d.Sources["foo"].URL.Path)
	assert.Equal(t, "blah", d.Sources["foo"].header["Accept"][0])
}

func TestParseSourceNoAlias(t *testing.T) {
	s, err := parseSource("foo.json")
	assert.NoError(t, err)
	assert.Equal(t, "foo", s.Alias)

	_, err = parseSource("../foo.json")
	assert.Error(t, err)

	_, err = parseSource("ftp://example.com/foo.yml")
	assert.Error(t, err)
}

func TestParseSourceWithAlias(t *testing.T) {
	s, err := parseSource("data=foo.json")
	assert.NoError(t, err)
	assert.Equal(t, "data", s.Alias)
	assert.Equal(t, "file", s.URL.Scheme)
	assert.True(t, s.URL.IsAbs())

	s, err = parseSource("data=/otherdir/foo.json")
	assert.NoError(t, err)
	assert.Equal(t, "data", s.Alias)
	assert.Equal(t, "file", s.URL.Scheme)
	assert.True(t, s.URL.IsAbs())
	assert.Equal(t, "/otherdir/foo.json", s.URL.Path)

	if runtime.GOOS == osWindows {
		s, err = parseSource("data=foo.json")
		assert.NoError(t, err)
		assert.Equalf(t, byte(':'), s.URL.Path[1], "Path was %s", s.URL.Path)

		s, err = parseSource(`data=\otherdir\foo.json`)
		assert.NoError(t, err)
		assert.Equal(t, "data", s.Alias)
		assert.Equal(t, "file", s.URL.Scheme)
		assert.True(t, s.URL.IsAbs())
		assert.Equal(t, `/otherdir/foo.json`, s.URL.Path)

		s, err = parseSource("data=C:\\windowsdir\\foo.json")
		assert.NoError(t, err)
		assert.Equal(t, "data", s.Alias)
		assert.Equal(t, "file", s.URL.Scheme)
		assert.True(t, s.URL.IsAbs())
		assert.Equal(t, "C:/windowsdir/foo.json", s.URL.Path)

		s, err = parseSource("data=\\\\somehost\\share\\foo.json")
		assert.NoError(t, err)
		assert.Equal(t, "data", s.Alias)
		assert.Equal(t, "file", s.URL.Scheme)
		assert.Equal(t, "somehost", s.URL.Host)
		assert.True(t, s.URL.IsAbs())
		assert.Equal(t, "/share/foo.json", s.URL.Path)
	}

	s, err = parseSource("data=sftp://example.com/blahblah/foo.json")
	assert.NoError(t, err)
	assert.Equal(t, "data", s.Alias)
	assert.Equal(t, "sftp", s.URL.Scheme)
	assert.True(t, s.URL.IsAbs())
	assert.Equal(t, "/blahblah/foo.json", s.URL.Path)

	s, err = parseSource("merged=merge:./foo.yaml|http://example.com/bar.json%3Ffoo=bar")
	assert.NoError(t, err)
	assert.Equal(t, "merged", s.Alias)
	assert.Equal(t, "merge", s.URL.Scheme)
	assert.Equal(t, "./foo.yaml|http://example.com/bar.json%3Ffoo=bar", s.URL.Opaque)
}

func TestDatasource(t *testing.T) {
	setup := func(ext, mime string, contents []byte) *Data {
		fname := "foo." + ext
		fs := afero.NewMemMapFs()
		var uPath string
		var f afero.File
		if runtime.GOOS == osWindows {
			_ = fs.Mkdir("C:\\tmp", 0777)
			f, _ = fs.Create("C:\\tmp\\" + fname)
			uPath = "C:/tmp/" + fname
		} else {
			_ = fs.Mkdir("/tmp", 0777)
			f, _ = fs.Create("/tmp/" + fname)
			uPath = "/tmp/" + fname
		}
		_, _ = f.Write(contents)

		sources := map[string]*Source{
			"foo": {
				Alias:     "foo",
				URL:       &url.URL{Scheme: "file", Path: uPath},
				mediaType: mime,
				fs:        fs,
			},
		}
		return &Data{Sources: sources}
	}
	test := func(ext, mime string, contents []byte) {
		data := setup(ext, mime, contents)
		expected := map[string]interface{}{"hello": map[string]interface{}{"cruel": "world"}}
		actual, err := data.Datasource("foo")
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
	fname := "foo.json"
	fs := afero.NewMemMapFs()
	var uPath string
	var f afero.File
	if runtime.GOOS == osWindows {
		_ = fs.Mkdir("C:\\tmp", 0777)
		f, _ = fs.Create("C:\\tmp\\" + fname)
		uPath = "C:/tmp/" + fname
	} else {
		_ = fs.Mkdir("/tmp", 0777)
		f, _ = fs.Create("/tmp/" + fname)
		uPath = "/tmp/" + fname
	}
	_, _ = f.Write([]byte("{}"))

	sources := map[string]*Source{
		"foo": {
			Alias:     "foo",
			URL:       &url.URL{Scheme: "file", Path: uPath},
			mediaType: jsonMimetype,
			fs:        fs,
		},
		"bar": {
			Alias: "bar",
			URL:   &url.URL{Scheme: "file", Path: "/bogus"},
			fs:    fs,
		},
	}
	data := &Data{Sources: sources}

	assert.True(t, data.DatasourceReachable("foo"))
	assert.False(t, data.DatasourceReachable("bar"))
}

func TestDatasourceExists(t *testing.T) {
	sources := map[string]*Source{
		"foo": {Alias: "foo"},
	}
	data := &Data{Sources: sources}
	assert.True(t, data.DatasourceExists("foo"))
	assert.False(t, data.DatasourceExists("bar"))
}

func TestInclude(t *testing.T) {
	ext := "txt"
	contents := "hello world"
	fname := "foo." + ext
	fs := afero.NewMemMapFs()

	var uPath string
	var f afero.File
	if runtime.GOOS == osWindows {
		_ = fs.Mkdir("C:\\tmp", 0777)
		f, _ = fs.Create("C:\\tmp\\" + fname)
		uPath = "C:/tmp/" + fname
	} else {
		_ = fs.Mkdir("/tmp", 0777)
		f, _ = fs.Create("/tmp/" + fname)
		uPath = "/tmp/" + fname
	}
	_, _ = f.Write([]byte(contents))

	sources := map[string]*Source{
		"foo": {
			Alias:     "foo",
			URL:       &url.URL{Scheme: "file", Path: uPath},
			mediaType: textMimetype,
			fs:        fs,
		},
	}
	data := &Data{
		Sources: sources,
	}
	actual, err := data.Include("foo")
	assert.NoError(t, err)
	assert.Equal(t, contents, actual)
}

type errorReader struct{}

func (e errorReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("error")
}

func TestReadStdin(t *testing.T) {
	defer func() {
		stdin = nil
	}()
	stdin = strings.NewReader("foo")
	out, err := readStdin(nil)
	assert.NoError(t, err)
	assert.Equal(t, []byte("foo"), out)

	stdin = errorReader{}
	_, err = readStdin(nil)
	assert.Error(t, err)
}

// nolint: megacheck
func TestDefineDatasource(t *testing.T) {
	d := &Data{}
	_, err := d.DefineDatasource("", "foo.json")
	assert.Error(t, err)

	d = &Data{}
	_, err = d.DefineDatasource("", "../foo.json")
	assert.Error(t, err)

	d = &Data{}
	_, err = d.DefineDatasource("", "ftp://example.com/foo.yml")
	assert.Error(t, err)

	d = &Data{}
	_, err = d.DefineDatasource("data", "foo.json")
	s := d.Sources["data"]
	assert.NoError(t, err)
	assert.Equal(t, "data", s.Alias)
	assert.Equal(t, "file", s.URL.Scheme)
	assert.True(t, s.URL.IsAbs())

	d = &Data{}
	_, err = d.DefineDatasource("data", "/otherdir/foo.json")
	s = d.Sources["data"]
	assert.NoError(t, err)
	assert.Equal(t, "data", s.Alias)
	assert.Equal(t, "file", s.URL.Scheme)
	assert.True(t, s.URL.IsAbs())
	assert.Equal(t, "/otherdir/foo.json", s.URL.Path)

	d = &Data{}
	_, err = d.DefineDatasource("data", "sftp://example.com/blahblah/foo.json")
	s = d.Sources["data"]
	assert.NoError(t, err)
	assert.Equal(t, "data", s.Alias)
	assert.Equal(t, "sftp", s.URL.Scheme)
	assert.True(t, s.URL.IsAbs())
	assert.Equal(t, "/blahblah/foo.json", s.URL.Path)

	d = &Data{
		Sources: map[string]*Source{
			"data": {Alias: "data"},
		},
	}
	_, err = d.DefineDatasource("data", "/otherdir/foo.json")
	s = d.Sources["data"]
	assert.NoError(t, err)
	assert.Equal(t, "data", s.Alias)
	assert.Nil(t, s.URL)
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
