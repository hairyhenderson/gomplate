package data

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"runtime"
	"testing"

	"github.com/hairyhenderson/gomplate/v3/internal/config"
	"github.com/spf13/afero"

	"github.com/stretchr/testify/assert"
)

const osWindows = "windows"

func TestNewData(t *testing.T) {
	d, err := NewData(nil, nil)
	assert.NoError(t, err)
	assert.Len(t, d.ds, 0)

	d, err = NewData([]string{"foo=http:///foo.json"}, nil)
	assert.NoError(t, err)
	assert.Equal(t, "/foo.json", d.ds["foo"].URL.Path)

	d, err = NewData([]string{"foo=http:///foo.json"}, []string{})
	assert.NoError(t, err)
	assert.Equal(t, "/foo.json", d.ds["foo"].URL.Path)
	assert.Empty(t, d.ds["foo"].Header)

	d, err = NewData([]string{"foo=http:///foo.json"}, []string{"bar=Accept: blah"})
	assert.NoError(t, err)
	assert.Equal(t, "/foo.json", d.ds["foo"].URL.Path)
	assert.Empty(t, d.ds["foo"].Header)
	assert.Equal(t, http.Header{"Accept": []string{"blah"}}, d.extraHeaders["bar"])

	d, err = NewData([]string{"foo=http:///foo.json"}, []string{"foo=Accept: blah"})
	assert.NoError(t, err)
	assert.Equal(t, "/foo.json", d.ds["foo"].URL.Path)
	assert.EqualValues(t, []string{"blah"}, d.ds["foo"].Header["Accept"])
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

		sources := map[string]config.DataSource{
			"foo": {
				URL: &url.URL{Scheme: "file", Path: uPath},
			},
		}
		ctx := config.WithFileSystem(context.Background(), fs)
		return &Data{ds: sources, ctx: ctx}
	}
	test := func(ext, mime string, contents []byte, expected interface{}) {
		data := setup(ext, mime, contents)

		actual, err := data.Datasource("foo")
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	}

	testObj := func(ext, mime string, contents []byte) {
		test(ext, mime, contents,
			map[string]interface{}{
				"hello": map[string]interface{}{"cruel": "world"},
			})
	}

	testObj("json", jsonMimetype, []byte(`{"hello":{"cruel":"world"}}`))
	testObj("yml", yamlMimetype, []byte("hello:\n  cruel: world\n"))
	test("json", jsonMimetype, []byte(`[1, "two", true]`),
		[]interface{}{1, "two", true})
	test("yaml", yamlMimetype, []byte("---\n- 1\n- two\n- true\n"),
		[]interface{}{1, "two", true})

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

	sources := map[string]config.DataSource{
		"foo": {
			URL: &url.URL{Scheme: "file", Path: uPath},
		},
		"bar": {
			URL: &url.URL{Scheme: "file", Path: "/bogus"},
		},
	}
	ctx := config.WithFileSystem(context.Background(), fs)
	data := &Data{ds: sources, ctx: ctx}

	assert.True(t, data.DatasourceReachable("foo"))
	assert.False(t, data.DatasourceReachable("bar"))
}

func TestDatasourceExists(t *testing.T) {
	sources := map[string]config.DataSource{
		"foo": {},
	}
	ctx := context.Background()
	data := &Data{ds: sources, ctx: ctx}
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

	sources := map[string]config.DataSource{
		"foo": {
			URL: &url.URL{Scheme: "file", Path: uPath},
		},
	}
	ctx := config.WithFileSystem(context.Background(), fs)
	data := &Data{ds: sources, ctx: ctx}
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
	ctx := context.Background()
	d := &Data{ctx: ctx}
	_, err := d.DefineDatasource("", "foo.json")
	assert.Error(t, err)

	d = &Data{ctx: ctx}
	_, err = d.DefineDatasource("", "../foo.json")
	assert.Error(t, err)

	d = &Data{ctx: ctx}
	_, err = d.DefineDatasource("", "ftp://example.com/foo.yml")
	assert.Error(t, err)

	d = &Data{ctx: ctx}
	_, err = d.DefineDatasource("data", "foo.json")
	s := d.ds["data"]
	assert.NoError(t, err)
	assert.Equal(t, "file", s.URL.Scheme)
	assert.True(t, s.URL.IsAbs())

	d = &Data{ctx: ctx}
	_, err = d.DefineDatasource("data", "/otherdir/foo.json")
	s = d.ds["data"]
	assert.NoError(t, err)
	assert.Equal(t, "file", s.URL.Scheme)
	assert.True(t, s.URL.IsAbs())
	assert.Equal(t, "/otherdir/foo.json", s.URL.Path)

	d = &Data{ctx: ctx}
	_, err = d.DefineDatasource("data", "sftp://example.com/blahblah/foo.json")
	s = d.ds["data"]
	assert.NoError(t, err)
	assert.Equal(t, "sftp", s.URL.Scheme)
	assert.True(t, s.URL.IsAbs())
	assert.Equal(t, "/blahblah/foo.json", s.URL.Path)

	d = &Data{
		ctx: ctx,
		ds:  map[string]config.DataSource{"data": {}},
	}
	_, err = d.DefineDatasource("data", "/otherdir/foo.json")
	s = d.ds["data"]
	assert.NoError(t, err)
	assert.Nil(t, s.URL)

	d = &Data{ctx: ctx}
	_, err = d.DefineDatasource("data", "/otherdir/foo?type=application/x-env")
	s = d.ds["data"]
	assert.NoError(t, err)
}

func TestFromConfig(t *testing.T) {
	ctx := context.Background()
	cfg := &config.Config{}
	expected := &Data{
		ctx: ctx,
		ds:  map[string]config.DataSource{},
	}
	assert.EqualValues(t, expected, FromConfig(ctx, cfg))

	cfg = &config.Config{
		DataSources: map[string]config.DataSource{
			"foo": {
				URL: mustParseURL("http://example.com"),
			},
		},
	}
	expected = &Data{
		ctx: ctx,
		ds: map[string]config.DataSource{
			"foo": {
				URL: mustParseURL("http://example.com"),
			},
		},
	}
	assert.EqualValues(t, expected, FromConfig(ctx, cfg))

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
	expected = &Data{
		ctx: ctx,
		ds: map[string]config.DataSource{
			"foo": {
				URL: mustParseURL("http://foo.com"),
			},
			"bar": {
				URL: mustParseURL("http://bar.com"),
				Header: http.Header{
					"Foo": []string{"bar"},
				},
			},
		},
		extraHeaders: map[string]http.Header{
			"baz": {
				"Foo": []string{"bar"},
			},
		},
	}
	assert.EqualValues(t, expected, FromConfig(ctx, cfg))
}
