package funcs

import (
	"context"
	"net/url"
	"runtime"
	"strconv"
	"testing"
	"testing/fstest"

	"github.com/hairyhenderson/gomplate/v4/internal/config"
	"github.com/hairyhenderson/gomplate/v4/internal/datafs"
	"github.com/hairyhenderson/gomplate/v4/internal/iohelpers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateDataSourceFuncs(t *testing.T) {
	t.Parallel()

	for i := range 10 {
		// Run this a bunch to catch race conditions
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			fmap := CreateDataSourceFuncs(ctx, nil)
			actual := fmap["_datasource"].(func() any)

			assert.Equal(t, ctx, actual().(*dataSourceFuncs).ctx)
		})
	}
}

const osWindows = "windows"

func TestDatasource(t *testing.T) {
	setup := func(ext string, contents []byte) *dataSourceFuncs {
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

		reg := datafs.NewRegistry()
		reg.Register("foo", config.DataSource{URL: &url.URL{Scheme: "file", Path: uPath}})

		d := &dataSourceFuncs{sr: datafs.NewSourceReader(reg), ctx: ctx}
		return d
	}

	test := func(ext, mime string, contents []byte, expected any) {
		data := setup(ext, contents)

		actual, err := data.Datasource("foo", "?type="+mime)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
	}

	testObj := func(ext, mime string, contents []byte) {
		test(ext, mime, contents,
			map[string]any{
				"hello": map[string]any{"cruel": "world"},
			})
	}

	testObj("json", iohelpers.JSONMimetype, []byte(`{"hello":{"cruel":"world"}}`))
	testObj("yml", iohelpers.YAMLMimetype, []byte("hello:\n  cruel: world\n"))
	test("json", iohelpers.JSONMimetype, []byte(`[1, "two", true]`),
		[]any{1, "two", true})
	test("yaml", iohelpers.YAMLMimetype, []byte("---\n- 1\n- two\n- true\n"),
		[]any{1, "two", true})

	d := setup("", nil)
	actual, err := d.Datasource("foo")
	require.NoError(t, err)
	assert.Empty(t, actual)

	_, err = d.Datasource("bar")
	require.Error(t, err)
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

	reg := datafs.NewRegistry()
	reg.Register("foo", config.DataSource{URL: &url.URL{Scheme: "file", Path: uPath}})
	reg.Register("bar", config.DataSource{URL: &url.URL{Scheme: "file", Path: "/bogus"}})

	data := &dataSourceFuncs{sr: datafs.NewSourceReader(reg)}
	data.ctx = ctx

	assert.True(t, data.DatasourceReachable("foo"))
	assert.False(t, data.DatasourceReachable("bar"))
}

func TestDatasourceExists(t *testing.T) {
	reg := datafs.NewRegistry()
	reg.Register("foo", config.DataSource{})
	data := &dataSourceFuncs{sr: datafs.NewSourceReader(reg)}

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

	reg := datafs.NewRegistry()
	reg.Register("foo", config.DataSource{URL: &url.URL{Scheme: "file", Path: uPath}})

	data := &dataSourceFuncs{sr: datafs.NewSourceReader(reg), ctx: ctx}

	actual, err := data.Include("foo")
	require.NoError(t, err)
	assert.Equal(t, contents, actual)
}

func TestDefineDatasource(t *testing.T) {
	reg := datafs.NewRegistry()
	d := &dataSourceFuncs{sr: datafs.NewSourceReader(reg)}
	_, err := d.DefineDatasource("", "foo.json")
	require.Error(t, err)

	d = &dataSourceFuncs{sr: datafs.NewSourceReader(reg)}
	_, err = d.DefineDatasource("", "../foo.json")
	require.Error(t, err)

	d = &dataSourceFuncs{sr: datafs.NewSourceReader(reg)}
	_, err = d.DefineDatasource("", "ftp://example.com/foo.yml")
	require.Error(t, err)

	reg = datafs.NewRegistry()
	d = &dataSourceFuncs{sr: datafs.NewSourceReader(reg)}
	_, err = d.DefineDatasource("data", "foo.json")
	s, _ := reg.Lookup("data")
	require.NoError(t, err)
	assert.Equal(t, &url.URL{Path: "foo.json"}, s.URL)

	reg = datafs.NewRegistry()
	d = &dataSourceFuncs{sr: datafs.NewSourceReader(reg)}
	_, err = d.DefineDatasource("data", "/otherdir/foo.json")
	s, _ = reg.Lookup("data")
	require.NoError(t, err)
	assert.Equal(t, "file", s.URL.Scheme)
	assert.True(t, s.URL.IsAbs())
	assert.Equal(t, "/otherdir/foo.json", s.URL.Path)

	reg = datafs.NewRegistry()
	d = &dataSourceFuncs{sr: datafs.NewSourceReader(reg)}
	_, err = d.DefineDatasource("data", "sftp://example.com/blahblah/foo.json")
	s, _ = reg.Lookup("data")
	require.NoError(t, err)
	assert.Equal(t, "sftp", s.URL.Scheme)
	assert.True(t, s.URL.IsAbs())
	assert.Equal(t, "/blahblah/foo.json", s.URL.Path)

	reg = datafs.NewRegistry()
	reg.Register("data", config.DataSource{})
	d = &dataSourceFuncs{sr: datafs.NewSourceReader(reg)}
	_, err = d.DefineDatasource("data", "/otherdir/foo.json")
	s, _ = reg.Lookup("data")
	require.NoError(t, err)
	assert.Nil(t, s.URL)

	reg = datafs.NewRegistry()
	d = &dataSourceFuncs{sr: datafs.NewSourceReader(reg)}
	_, err = d.DefineDatasource("data", "/otherdir/foo?type=application/x-env")
	require.NoError(t, err)

	s, _ = reg.Lookup("data")
	require.NotNil(t, s)
	assert.Equal(t, "/otherdir/foo", s.URL.Path)
}

func TestListDatasources(t *testing.T) {
	reg := datafs.NewRegistry()
	reg.Register("foo", config.DataSource{})
	reg.Register("bar", config.DataSource{})

	d := &dataSourceFuncs{sr: datafs.NewSourceReader(reg)}

	assert.Equal(t, []string{"bar", "foo"}, d.ListDatasources())
}
