package data

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"runtime"
	"testing"
	"testing/fstest"

	"github.com/hairyhenderson/go-fsimpl"
	"github.com/hairyhenderson/go-fsimpl/httpfs"
	"github.com/hairyhenderson/gomplate/v4/internal/config"
	"github.com/hairyhenderson/gomplate/v4/internal/datafs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const osWindows = "windows"

func mustParseURL(in string) *url.URL {
	u, _ := url.Parse(in)
	return u
}

func TestNewData(t *testing.T) {
	d, err := NewData(nil, nil)
	require.NoError(t, err)
	assert.Len(t, d.Sources, 0)

	d, err = NewData([]string{"foo=http:///foo.json"}, nil)
	require.NoError(t, err)
	assert.Equal(t, "/foo.json", d.Sources["foo"].URL.Path)

	d, err = NewData([]string{"foo=http:///foo.json"}, []string{})
	require.NoError(t, err)
	assert.Equal(t, "/foo.json", d.Sources["foo"].URL.Path)
	assert.Empty(t, d.Sources["foo"].Header)

	d, err = NewData([]string{"foo=http:///foo.json"}, []string{"bar=Accept: blah"})
	require.NoError(t, err)
	assert.Equal(t, "/foo.json", d.Sources["foo"].URL.Path)
	assert.Empty(t, d.Sources["foo"].Header)

	d, err = NewData([]string{"foo=http:///foo.json"}, []string{"foo=Accept: blah"})
	require.NoError(t, err)
	assert.Equal(t, "/foo.json", d.Sources["foo"].URL.Path)
	assert.Equal(t, "blah", d.Sources["foo"].Header["Accept"][0])
}

func TestDatasource(t *testing.T) {
	setup := func(ext string, contents []byte) *Data {
		fname := "foo." + ext
		var uPath string
		if runtime.GOOS == osWindows {
			uPath = "C:/tmp/" + fname
		} else {
			uPath = "/tmp/" + fname
		}

		fsys := datafs.WrapWdFS(fstest.MapFS{
			"tmp/" + fname: &fstest.MapFile{Data: contents},
		})
		ctx := datafs.ContextWithFSProvider(context.Background(), datafs.WrappedFSProvider(fsys, "file", ""))

		sources := map[string]config.DataSource{
			"foo": {
				URL: &url.URL{Scheme: "file", Path: uPath},
			},
		}
		return &Data{Sources: sources, Ctx: ctx}
	}

	test := func(ext, mime string, contents []byte, expected interface{}) {
		data := setup(ext, contents)

		actual, err := data.Datasource("foo", "?type="+mime)
		require.NoError(t, err)
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

	d := setup("", nil)
	actual, err := d.Datasource("foo")
	require.NoError(t, err)
	assert.Equal(t, "", actual)

	_, err = d.Datasource("bar")
	assert.Error(t, err)
}

func TestDatasourceReachable(t *testing.T) {
	fname := "foo.json"
	var uPath string
	if runtime.GOOS == osWindows {
		uPath = "C:/tmp/" + fname
	} else {
		uPath = "/tmp/" + fname
	}

	fsys := datafs.WrapWdFS(fstest.MapFS{
		"tmp/" + fname: &fstest.MapFile{Data: []byte("{}")},
	})
	ctx := datafs.ContextWithFSProvider(context.Background(), datafs.WrappedFSProvider(fsys, "file", ""))

	sources := map[string]config.DataSource{
		"foo": {
			URL: &url.URL{Scheme: "file", Path: uPath},
		},
		"bar": {
			URL: &url.URL{Scheme: "file", Path: "/bogus"},
		},
	}
	data := &Data{Sources: sources, Ctx: ctx}

	assert.True(t, data.DatasourceReachable("foo"))
	assert.False(t, data.DatasourceReachable("bar"))
}

func TestDatasourceExists(t *testing.T) {
	sources := map[string]config.DataSource{
		"foo": {},
	}
	data := &Data{Sources: sources}
	assert.True(t, data.DatasourceExists("foo"))
	assert.False(t, data.DatasourceExists("bar"))
}

func TestInclude(t *testing.T) {
	ext := "txt"
	contents := "hello world"
	fname := "foo." + ext

	var uPath string
	if runtime.GOOS == osWindows {
		uPath = "C:/tmp/" + fname
	} else {
		uPath = "/tmp/" + fname
	}

	fsys := datafs.WrapWdFS(fstest.MapFS{
		"tmp/" + fname: &fstest.MapFile{Data: []byte(contents)},
	})
	ctx := datafs.ContextWithFSProvider(context.Background(), datafs.WrappedFSProvider(fsys, "file", ""))

	sources := map[string]config.DataSource{
		"foo": {
			URL: &url.URL{Scheme: "file", Path: uPath},
		},
	}
	data := &Data{Sources: sources, Ctx: ctx}
	actual, err := data.Include("foo")
	require.NoError(t, err)
	assert.Equal(t, contents, actual)
}

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
	require.NoError(t, err)
	assert.EqualValues(t, &url.URL{Path: "foo.json"}, s.URL)

	d = &Data{}
	_, err = d.DefineDatasource("data", "/otherdir/foo.json")
	s = d.Sources["data"]
	require.NoError(t, err)
	assert.Equal(t, "file", s.URL.Scheme)
	assert.True(t, s.URL.IsAbs())
	assert.Equal(t, "/otherdir/foo.json", s.URL.Path)

	d = &Data{}
	_, err = d.DefineDatasource("data", "sftp://example.com/blahblah/foo.json")
	s = d.Sources["data"]
	require.NoError(t, err)
	assert.Equal(t, "sftp", s.URL.Scheme)
	assert.True(t, s.URL.IsAbs())
	assert.Equal(t, "/blahblah/foo.json", s.URL.Path)

	d = &Data{
		Sources: map[string]config.DataSource{
			"data": {},
		},
	}
	_, err = d.DefineDatasource("data", "/otherdir/foo.json")
	s = d.Sources["data"]
	require.NoError(t, err)
	assert.Nil(t, s.URL)

	d = &Data{}
	_, err = d.DefineDatasource("data", "/otherdir/foo?type=application/x-env")
	require.NoError(t, err)
	s = d.Sources["data"]
	require.NotNil(t, s)
	assert.Equal(t, "/otherdir/foo", s.URL.Path)
}

func TestFromConfig(t *testing.T) {
	ctx := context.Background()

	cfg := &config.Config{}
	actual := FromConfig(ctx, cfg)
	expected := &Data{
		Ctx:     actual.Ctx,
		Sources: map[string]config.DataSource{},
	}
	assert.EqualValues(t, expected, actual)

	cfg = &config.Config{
		DataSources: map[string]config.DataSource{
			"foo": {
				URL: mustParseURL("http://example.com"),
			},
		},
	}
	actual = FromConfig(ctx, cfg)
	expected = &Data{
		Ctx: actual.Ctx,
		Sources: map[string]config.DataSource{
			"foo": {
				URL: mustParseURL("http://example.com"),
			},
		},
	}
	assert.EqualValues(t, expected, actual)

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
	actual = FromConfig(ctx, cfg)
	expected = &Data{
		Ctx: actual.Ctx,
		Sources: map[string]config.DataSource{
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
		ExtraHeaders: map[string]http.Header{
			"baz": {
				"Foo": []string{"bar"},
			},
		},
	}
	assert.EqualValues(t, expected, actual)
}

func TestListDatasources(t *testing.T) {
	sources := map[string]config.DataSource{
		"foo": {},
		"bar": {},
	}
	data := &Data{Sources: sources}

	assert.Equal(t, []string{"bar", "foo"}, data.ListDatasources())
}

func TestResolveURL(t *testing.T) {
	out, err := resolveURL(mustParseURL("http://example.com/foo.json"), "bar.json")
	assert.NoError(t, err)
	assert.Equal(t, "http://example.com/bar.json", out.String())

	out, err = resolveURL(mustParseURL("http://example.com/a/b/?n=2"), "bar.json?q=1")
	assert.NoError(t, err)
	assert.Equal(t, "http://example.com/a/b/bar.json?n=2&q=1", out.String())

	out, err = resolveURL(mustParseURL("git+file:///tmp/myrepo"), "//myfile?type=application/json")
	assert.NoError(t, err)
	assert.Equal(t, "git+file:///tmp/myrepo//myfile?type=application/json", out.String())

	out, err = resolveURL(mustParseURL("git+file:///tmp/foo/bar/"), "//myfile?type=application/json")
	assert.NoError(t, err)
	assert.Equal(t, "git+file:///tmp/foo/bar//myfile?type=application/json", out.String())

	out, err = resolveURL(mustParseURL("git+file:///tmp/myrepo/"), ".//myfile?type=application/json")
	assert.NoError(t, err)
	assert.Equal(t, "git+file:///tmp/myrepo//myfile?type=application/json", out.String())

	out, err = resolveURL(mustParseURL("git+file:///tmp/repo//foo.txt"), "")
	assert.NoError(t, err)
	assert.Equal(t, "git+file:///tmp/repo//foo.txt", out.String())

	out, err = resolveURL(mustParseURL("git+file:///tmp/myrepo"), ".//myfile?type=application/json")
	assert.NoError(t, err)
	assert.Equal(t, "git+file:///tmp/myrepo//myfile?type=application/json", out.String())

	out, err = resolveURL(mustParseURL("git+file:///tmp/myrepo//foo/?type=application/json"), "bar/myfile")
	assert.NoError(t, err)
	// note that the '/' in the query string is encoded to %2F - that's OK
	assert.Equal(t, "git+file:///tmp/myrepo//foo/bar/myfile?type=application%2Fjson", out.String())

	// both base and relative may not contain "//"
	_, err = resolveURL(mustParseURL("git+ssh://git@example.com/foo//bar"), ".//myfile")
	assert.Error(t, err)

	_, err = resolveURL(mustParseURL("git+ssh://git@example.com/foo//bar"), "baz//myfile")
	assert.Error(t, err)

	// relative urls must remain relative
	out, err = resolveURL(mustParseURL("tmp/foo.json"), "")
	require.NoError(t, err)
	assert.Equal(t, "tmp/foo.json", out.String())
}

func TestReadFileContent(t *testing.T) {
	wd, _ := os.Getwd()
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})
	_ = os.Chdir("/")

	mux := http.NewServeMux()
	mux.HandleFunc("/foo.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", jsonMimetype)
		w.Write([]byte(`{"foo": "bar"}`))
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	fsys := datafs.WrapWdFS(fstest.MapFS{
		"foo.json":          &fstest.MapFile{Data: []byte(`{"foo": "bar"}`)},
		"dir/1.yaml":        &fstest.MapFile{Data: []byte(`foo: bar`)},
		"dir/2.yaml":        &fstest.MapFile{Data: []byte(`baz: qux`)},
		"dir/sub/sub1.yaml": &fstest.MapFile{Data: []byte(`quux: corge`)},
	})

	fsp := fsimpl.NewMux()
	fsp.Add(httpfs.FS)
	fsp.Add(datafs.WrappedFSProvider(fsys, "file", ""))

	ctx := datafs.ContextWithFSProvider(context.Background(), fsp)

	d := Data{}

	fc, err := d.readFileContent(ctx, mustParseURL("file:///foo.json"), nil)
	require.NoError(t, err)
	assert.Equal(t, []byte(`{"foo": "bar"}`), fc.b)

	fc, err = d.readFileContent(ctx, mustParseURL("dir/"), nil)
	require.NoError(t, err)
	assert.JSONEq(t, `["1.yaml", "2.yaml", "sub"]`, string(fc.b))

	fc, err = d.readFileContent(ctx, mustParseURL(srv.URL+"/foo.json"), nil)
	require.NoError(t, err)
	assert.Equal(t, []byte(`{"foo": "bar"}`), fc.b)
}
