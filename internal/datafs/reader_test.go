package datafs

import (
	"context"
	"fmt"
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
	"github.com/hairyhenderson/gomplate/v4/internal/iohelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const osWindows = "windows"

func TestResolveURL(t *testing.T) {
	out, err := resolveURL(*mustParseURL("http://example.com/foo.json"), "bar.json")
	require.NoError(t, err)
	assert.Equal(t, "http://example.com/bar.json", out.String())

	out, err = resolveURL(*mustParseURL("http://example.com/a/b/?n=2"), "bar.json?q=1")
	require.NoError(t, err)
	assert.Equal(t, "http://example.com/a/b/bar.json?n=2&q=1", out.String())

	out, err = resolveURL(*mustParseURL("git+file:///tmp/myrepo"), "//myfile?type=application/json")
	require.NoError(t, err)
	assert.Equal(t, "git+file:///tmp/myrepo//myfile?type=application/json", out.String())

	out, err = resolveURL(*mustParseURL("git+file:///tmp/foo/bar/"), "//myfile?type=application/json")
	require.NoError(t, err)
	assert.Equal(t, "git+file:///tmp/foo/bar//myfile?type=application/json", out.String())

	out, err = resolveURL(*mustParseURL("git+file:///tmp/myrepo/"), ".//myfile?type=application/json")
	require.NoError(t, err)
	assert.Equal(t, "git+file:///tmp/myrepo//myfile?type=application/json", out.String())

	out, err = resolveURL(*mustParseURL("git+file:///tmp/repo//foo.txt"), "")
	require.NoError(t, err)
	assert.Equal(t, "git+file:///tmp/repo//foo.txt", out.String())

	out, err = resolveURL(*mustParseURL("git+file:///tmp/myrepo"), ".//myfile?type=application/json")
	require.NoError(t, err)
	assert.Equal(t, "git+file:///tmp/myrepo//myfile?type=application/json", out.String())

	out, err = resolveURL(*mustParseURL("git+file:///tmp/myrepo//foo/?type=application/json"), "bar/myfile")
	require.NoError(t, err)
	// note that the '/' in the query string is encoded to %2F - that's OK
	assert.Equal(t, "git+file:///tmp/myrepo//foo/bar/myfile?type=application%2Fjson", out.String())

	// both base and relative may not contain "//"
	_, err = resolveURL(*mustParseURL("git+ssh://git@example.com/foo//bar"), ".//myfile")
	require.Error(t, err)

	_, err = resolveURL(*mustParseURL("git+ssh://git@example.com/foo//bar"), "baz//myfile")
	require.Error(t, err)

	// relative base URLs must remain relative
	out, err = resolveURL(*mustParseURL("tmp/foo.json"), "")
	require.NoError(t, err)
	assert.Equal(t, "tmp/foo.json", out.String())

	// relative implicit file URLs without volume or scheme are OK
	out, err = resolveURL(*mustParseURL("/tmp/"), "foo.json")
	require.NoError(t, err)
	assert.Equal(t, "tmp/foo.json", out.String())

	// relative base URLs in parent directories are OK
	out, err = resolveURL(*mustParseURL("../../tmp/foo.json"), "")
	require.NoError(t, err)
	assert.Equal(t, "../../tmp/foo.json", out.String())

	out, err = resolveURL(*mustParseURL("../../tmp/"), "sub/foo.json")
	require.NoError(t, err)
	assert.Equal(t, "../../tmp/sub/foo.json", out.String())

	t.Run("aws+sm", func(t *testing.T) {
		out, err = resolveURL(*mustParseURL("aws+sm:"), "foo")
		require.NoError(t, err)
		assert.Equal(t, "aws+sm:foo", out.String())

		out, err = resolveURL(*mustParseURL("aws+sm:foo/"), "bar")
		require.NoError(t, err)
		assert.Equal(t, "aws+sm:foo/bar", out.String())

		out, err = resolveURL(*mustParseURL("aws+sm:"), "/foo")
		require.NoError(t, err)
		assert.Equal(t, "aws+sm:///foo", out.String())

		out, err = resolveURL(*mustParseURL("aws+sm:///foo/"), "bar")
		require.NoError(t, err)
		assert.Equal(t, "aws+sm:///foo/bar", out.String())
	})
}

func BenchmarkResolveURL(b *testing.B) {
	args := []struct {
		url url.URL
		rel string
	}{
		{*mustParseURL("http://example.com/foo.json"), "bar.json"},
		{*mustParseURL("http://example.com/a/b/?n=2"), "bar.json?q=1"},
		{*mustParseURL("git+file:///tmp/myrepo"), "//myfile?type=application/json"},
		{*mustParseURL("git+file:///tmp/myrepo2"), ".//myfile?type=application/json"},
		{*mustParseURL("git+file:///tmp/foo/bar/"), "//myfile?type=application/json"},
		{*mustParseURL("git+file:///tmp/myrepo/"), ".//myfile?type=application/json"},
		{*mustParseURL("git+file:///tmp/repo//foo.txt"), ""},
		{*mustParseURL("git+file:///tmp/myrepo//foo/?type=application/json"), "bar/myfile"},
		{*mustParseURL("tmp/foo.json"), ""},
		{*mustParseURL("/tmp/"), "foo.json"},
		{*mustParseURL("../../tmp/foo.json"), ""},
		{*mustParseURL("../../tmp/"), "sub/foo.json"},
		{*mustParseURL("aws+sm:"), "foo"},
		{*mustParseURL("aws+sm:"), "/foo"},
		{*mustParseURL("aws+sm:foo"), "bar"},
		{*mustParseURL("aws+sm:///foo"), "bar"},
	}

	b.ResetTimer()

	for _, a := range args {
		b.Run(fmt.Sprintf("base=%s_rel=%s", &a.url, a.rel), func(b *testing.B) {
			for b.Loop() {
				_, _ = resolveURL(a.url, a.rel)
			}
		})
	}
}

func TestReadFileContent(t *testing.T) {
	wd, _ := os.Getwd()
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})
	_ = os.Chdir("/")

	mux := http.NewServeMux()
	mux.HandleFunc("/foo.json", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", iohelpers.JSONMimetype)
		w.Write([]byte(`{"foo": "bar"}`))
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	fsys := WrapWdFS(fstest.MapFS{
		"foo.json":          &fstest.MapFile{Data: []byte(`{"foo": "bar"}`)},
		"dir/1.yaml":        &fstest.MapFile{Data: []byte(`foo: bar`)},
		"dir/2.yaml":        &fstest.MapFile{Data: []byte(`baz: qux`)},
		"dir/sub/sub1.yaml": &fstest.MapFile{Data: []byte(`quux: corge`)},
	})

	fsp := fsimpl.NewMux()
	fsp.Add(httpfs.FS)
	fsp.Add(WrappedFSProvider(fsys, "file", ""))

	ctx := ContextWithFSProvider(context.Background(), fsp)

	reg := NewRegistry()
	sr := &dsReader{Registry: reg}

	fc, err := sr.readFileContent(ctx, mustParseURL("file:///foo.json"), nil)
	require.NoError(t, err)
	assert.JSONEq(t, `{"foo": "bar"}`, string(fc.b))

	fc, err = sr.readFileContent(ctx, mustParseURL("dir/"), nil)
	require.NoError(t, err)
	assert.JSONEq(t, `["1.yaml", "2.yaml", "sub"]`, string(fc.b))

	fc, err = sr.readFileContent(ctx, mustParseURL(srv.URL+"/foo.json"), nil)
	require.NoError(t, err)
	assert.JSONEq(t, `{"foo": "bar"}`, string(fc.b))
}

func TestDatasource(t *testing.T) {
	setup := func(ext string, contents []byte) (context.Context, *dsReader) {
		fname := "foo." + ext
		var uPath string
		if runtime.GOOS == osWindows {
			uPath = "C:/tmp/" + fname
		} else {
			uPath = "/tmp/" + fname
		}

		fsys := WrapWdFS(fstest.MapFS{
			"tmp/" + fname: &fstest.MapFile{Data: contents},
		})
		ctx := ContextWithFSProvider(context.Background(), WrappedFSProvider(fsys, "file", ""))

		reg := NewRegistry()
		reg.Register("foo", config.DataSource{URL: &url.URL{Scheme: "file", Path: uPath}})

		return ctx, &dsReader{Registry: reg}
	}

	test := func(ext, mime string, contents []byte) {
		ctx, data := setup(ext, contents)

		ct, b, err := data.ReadSource(ctx, "foo", "?type="+mime)
		require.NoError(t, err)
		assert.Equal(t, contents, b)
		assert.Equal(t, mime, ct)
	}

	testObj := func(ext, mime string, contents []byte) {
		test(ext, mime, contents)
	}

	testObj("json", iohelpers.JSONMimetype, []byte(`{"hello":{"cruel":"world"}}`))
	testObj("yml", iohelpers.YAMLMimetype, []byte("hello:\n  cruel: world\n"))
	test("json", iohelpers.JSONMimetype, []byte(`[1, "two", true]`))
	test("yaml", iohelpers.YAMLMimetype, []byte("---\n- 1\n- two\n- true\n"))

	ctx, d := setup("", nil)
	ct, b, err := d.ReadSource(ctx, "foo")
	require.NoError(t, err)
	assert.Empty(t, b)
	assert.Equal(t, iohelpers.TextMimetype, ct)

	_, _, err = d.ReadSource(ctx, "bar")
	require.Error(t, err)
}
