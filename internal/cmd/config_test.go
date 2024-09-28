package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"testing"
	"testing/fstest"
	"time"

	"github.com/hairyhenderson/go-fsimpl"
	"github.com/hairyhenderson/gomplate/v4"
	"github.com/hairyhenderson/gomplate/v4/internal/datafs"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadConfigFile(t *testing.T) {
	ctx := context.Background()
	fsys := fstest.MapFS{}
	ctx = datafs.ContextWithFSProvider(ctx, fsimpl.FSProviderFunc(func(_ *url.URL) (fs.FS, error) {
		return fsys, nil
	}))
	cmd := &cobra.Command{}

	_, err := readConfigFile(ctx, cmd)
	require.NoError(t, err)

	cmd.Flags().String("config", defaultConfigFile, "foo")

	_, err = readConfigFile(ctx, cmd)
	require.NoError(t, err)

	cmd.ParseFlags([]string{"--config", "config.file"})

	_, err = readConfigFile(ctx, cmd)
	require.Error(t, err)

	cmd = &cobra.Command{}
	cmd.Flags().String("config", defaultConfigFile, "foo")

	fsys[defaultConfigFile] = &fstest.MapFile{}

	cfg, err := readConfigFile(ctx, cmd)
	require.NoError(t, err)
	assert.EqualValues(t, &gomplate.Config{}, cfg)

	cmd.ParseFlags([]string{"--config", "config.yaml"})

	fsys["config.yaml"] = &fstest.MapFile{Data: []byte("in: hello world\n")}

	cfg, err = readConfigFile(ctx, cmd)
	require.NoError(t, err)
	assert.EqualValues(t, &gomplate.Config{Input: "hello world"}, cfg)

	fsys["config.yaml"] = &fstest.MapFile{Data: []byte("in: hello world\nin: \n")}

	_, err = readConfigFile(ctx, cmd)
	require.Error(t, err)
}

func TestLoadConfig(t *testing.T) {
	ctx := context.Background()
	fsys := fstest.MapFS{}
	ctx = datafs.ContextWithFSProvider(ctx, fsimpl.FSProviderFunc(func(_ *url.URL) (fs.FS, error) {
		return fsys, nil
	}))

	stdin, stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{}
	cmd := &cobra.Command{}
	cmd.SetIn(stdin)
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)

	cmd.Args = optionalExecArgs
	cmd.Flags().StringSlice("file", []string{"-"}, "...")
	cmd.Flags().StringSlice("out", []string{"-"}, "...")
	cmd.Flags().String("in", ".", "...")
	cmd.Flags().String("output-dir", ".", "...")
	cmd.Flags().String("left-delim", "{{", "...")
	cmd.Flags().String("right-delim", "}}", "...")
	cmd.Flags().Bool("exec-pipe", false, "...")
	cmd.ParseFlags(nil)

	out, err := loadConfig(ctx, cmd, cmd.Flags().Args())
	expected := &gomplate.Config{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
	}
	require.NoError(t, err)
	assert.EqualValues(t, expected, out)

	cmd.ParseFlags([]string{"--in", "foo"})
	out, err = loadConfig(ctx, cmd, cmd.Flags().Args())
	expected = &gomplate.Config{
		Input:  "foo",
		Stdin:  stdin,
		Stdout: out.Stdout,
		Stderr: stderr,
	}
	require.NoError(t, err)
	assert.EqualValues(t, expected, out)

	cmd.ParseFlags([]string{"--in", "foo", "--exec-pipe", "--", "tr", "[a-z]", "[A-Z]"})
	out, err = loadConfig(ctx, cmd, cmd.Flags().Args())
	expected = &gomplate.Config{
		Input:    "foo",
		ExecPipe: true,
		PostExec: []string{"tr", "[a-z]", "[A-Z]"},
		Stdin:    stdin,
		Stdout:   out.Stdout,
		Stderr:   stderr,
	}
	require.NoError(t, err)
	assert.EqualValues(t, expected, out)
}

func TestPostExecInput(t *testing.T) {
	t.Parallel()

	cfg := &gomplate.Config{ExecPipe: false}
	assert.Equal(t, os.Stdin, postExecInput(cfg))

	cfg = &gomplate.Config{ExecPipe: true}

	pipe := postExecInput(cfg)
	assert.IsType(t, &bytes.Buffer{}, pipe)
	assert.Equal(t, []string{"-"}, cfg.OutputFiles)
	assert.Equal(t, pipe, cfg.Stdout)

	stdin := &bytes.Buffer{}
	cfg = &gomplate.Config{ExecPipe: false, Stdin: stdin}
	pipe = postExecInput(cfg)
	assert.Equal(t, stdin, pipe)
}

func TestCobraConfig(t *testing.T) {
	t.Parallel()
	cmd := &cobra.Command{}
	cmd.Flags().StringSlice("file", []string{"-"}, "...")
	cmd.Flags().StringSlice("out", []string{"-"}, "...")
	cmd.Flags().String("output-dir", ".", "...")
	cmd.Flags().String("left-delim", "{{", "...")
	cmd.Flags().String("right-delim", "}}", "...")
	cmd.ParseFlags(nil)

	cfg, err := cobraConfig(cmd, cmd.Flags().Args())
	require.NoError(t, err)
	assert.EqualValues(t, &gomplate.Config{}, cfg)

	cmd.ParseFlags([]string{"--file", "in", "--", "echo", "foo"})

	cfg, err = cobraConfig(cmd, cmd.Flags().Args())
	require.NoError(t, err)
	assert.EqualValues(t, &gomplate.Config{
		InputFiles: []string{"in"},
		PostExec:   []string{"echo", "foo"},
	}, cfg)
}

func TestProcessIncludes(t *testing.T) {
	t.Parallel()
	data := []struct {
		inc, exc, expected []string
	}{
		{nil, nil, nil},
		{[]string{}, []string{}, nil},
		{nil, []string{"*.foo"}, []string{"*.foo"}},
		{[]string{"*.bar"}, []string{"a*.bar"}, []string{"*", "!*.bar", "a*.bar"}},
		{[]string{"*.bar"}, nil, []string{"*", "!*.bar"}},
	}

	for _, d := range data {
		assert.EqualValues(t, d.expected, processIncludes(d.inc, d.exc))
	}
}

func TestPickConfigFile(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("config", defaultConfigFile, "foo")

	t.Run("default", func(t *testing.T) {
		cf, req, skip := pickConfigFile(cmd)
		assert.False(t, req)
		assert.False(t, skip)
		assert.Equal(t, defaultConfigFile, cf)
	})

	t.Run("GOMPLATE_CONFIG env var", func(t *testing.T) {
		t.Setenv("GOMPLATE_CONFIG", "foo.yaml")
		cf, req, skip := pickConfigFile(cmd)
		assert.True(t, req)
		assert.False(t, skip)
		assert.Equal(t, "foo.yaml", cf)
	})

	t.Run("--config flag", func(t *testing.T) {
		cmd.ParseFlags([]string{"--config", "config.file"})
		cf, req, skip := pickConfigFile(cmd)
		assert.True(t, req)
		assert.False(t, skip)
		assert.Equal(t, "config.file", cf)

		t.Setenv("GOMPLATE_CONFIG", "ignored.yaml")
		cf, req, skip = pickConfigFile(cmd)
		assert.True(t, req)
		assert.False(t, skip)
		assert.Equal(t, "config.file", cf)
	})

	t.Run("--config flag with empty value should skip reading", func(t *testing.T) {
		cmd.ParseFlags([]string{"--config", ""})
		cf, req, skip := pickConfigFile(cmd)
		assert.False(t, req)
		assert.True(t, skip)
		assert.Equal(t, "", cf)
	})

	t.Run("GOMPLATE_CONFIG env var with empty value should skip reading", func(t *testing.T) {
		t.Setenv("GOMPLATE_CONFIG", "")
		cf, req, skip := pickConfigFile(cmd)
		assert.False(t, req)
		assert.True(t, skip)
		assert.Equal(t, "", cf)
	})
}

func TestApplyEnvVars(t *testing.T) {
	t.Run("invalid GOMPLATE_PLUGIN_TIMEOUT", func(t *testing.T) {
		t.Setenv("GOMPLATE_PLUGIN_TIMEOUT", "bogus")
		_, err := applyEnvVars(context.Background(), &gomplate.Config{})
		require.Error(t, err)
	})

	data := []struct {
		input, expected *gomplate.Config
		env             string
		value           string
	}{
		{
			&gomplate.Config{PluginTimeout: 2 * time.Second},
			&gomplate.Config{PluginTimeout: 2 * time.Second},
			"GOMPLATE_PLUGIN_TIMEOUT", "bogus",
		},
		{
			&gomplate.Config{},
			&gomplate.Config{PluginTimeout: 2 * time.Second},
			"GOMPLATE_PLUGIN_TIMEOUT", "2s",
		},
		{
			&gomplate.Config{PluginTimeout: 100 * time.Millisecond},
			&gomplate.Config{PluginTimeout: 100 * time.Millisecond},
			"GOMPLATE_PLUGIN_TIMEOUT", "2s",
		},
		{
			&gomplate.Config{},
			&gomplate.Config{Experimental: false},
			"GOMPLATE_EXPERIMENTAL", "bogus",
		},
		{
			&gomplate.Config{},
			&gomplate.Config{Experimental: true},
			"GOMPLATE_EXPERIMENTAL", "true",
		},
		{
			&gomplate.Config{Experimental: true},
			&gomplate.Config{Experimental: true},
			"GOMPLATE_EXPERIMENTAL", "false",
		},
		{
			&gomplate.Config{},
			&gomplate.Config{LDelim: "--"},
			"GOMPLATE_LEFT_DELIM", "--",
		},
		{
			&gomplate.Config{LDelim: "{{"},
			&gomplate.Config{LDelim: "{{"},
			"GOMPLATE_LEFT_DELIM", "--",
		},
		{
			&gomplate.Config{},
			&gomplate.Config{RDelim: ")>"},
			"GOMPLATE_RIGHT_DELIM", ")>",
		},
		{
			&gomplate.Config{RDelim: "}}"},
			&gomplate.Config{RDelim: "}}"},
			"GOMPLATE_RIGHT_DELIM", ")>",
		},
		{
			&gomplate.Config{RDelim: "}}"},
			&gomplate.Config{RDelim: "}}"},
			"GOMPLATE_RIGHT_DELIM", "",
		},
	}

	for i, d := range data {
		t.Run(fmt.Sprintf("applyEnvVars_%s_%s/%d", d.env, d.value, i), func(t *testing.T) {
			t.Setenv(d.env, d.value)

			actual, err := applyEnvVars(context.Background(), d.input)
			require.NoError(t, err)
			assert.EqualValues(t, d.expected, actual)
		})
	}
}

func mustURL(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}

	return u
}

func TestParseDataSourceFlags(t *testing.T) {
	t.Parallel()
	cfg := &gomplate.Config{}
	err := ParseDataSourceFlags(cfg, nil, nil, nil, nil)
	require.NoError(t, err)
	assert.EqualValues(t, &gomplate.Config{}, cfg)

	cfg = &gomplate.Config{}
	err = ParseDataSourceFlags(cfg, []string{"foo/bar/baz.json"}, nil, nil, nil)
	require.Error(t, err)

	cfg = &gomplate.Config{}
	err = ParseDataSourceFlags(cfg, []string{"baz=foo/bar/baz.json"}, nil, nil, nil)
	require.NoError(t, err)
	expected := &gomplate.Config{
		DataSources: map[string]gomplate.DataSource{
			"baz": {URL: mustURL("foo/bar/baz.json")},
		},
	}
	assert.EqualValues(t, expected, cfg, "expected: %+v\nactual: %+v\n", expected, cfg)

	cfg = &gomplate.Config{}
	err = ParseDataSourceFlags(cfg,
		[]string{"baz=foo/bar/baz.json"},
		nil,
		nil,
		[]string{"baz=Accept: application/json"})
	require.NoError(t, err)
	assert.EqualValues(t, &gomplate.Config{
		DataSources: map[string]gomplate.DataSource{
			"baz": {
				URL: mustURL("foo/bar/baz.json"),
				Header: http.Header{
					"Accept": {"application/json"},
				},
			},
		},
	}, cfg)

	cfg = &gomplate.Config{}
	err = ParseDataSourceFlags(cfg,
		[]string{"baz=foo/bar/baz.json"},
		[]string{"foo=http://example.com"},
		nil,
		[]string{
			"foo=Accept: application/json",
			"bar=Authorization: Basic xxxxx",
		},
	)
	require.NoError(t, err)
	assert.EqualValues(t, &gomplate.Config{
		DataSources: map[string]gomplate.DataSource{
			"baz": {URL: mustURL("foo/bar/baz.json")},
		},
		Context: map[string]gomplate.DataSource{
			"foo": {
				URL: mustURL("http://example.com"),
				Header: http.Header{
					"Accept": {"application/json"},
				},
			},
		},
		ExtraHeaders: map[string]http.Header{
			"bar": {"Authorization": {"Basic xxxxx"}},
		},
	}, cfg)

	cfg = &gomplate.Config{}
	err = ParseDataSourceFlags(cfg,
		nil,
		nil,
		[]string{"foo=http://example.com", "file.tmpl", "tmpldir/"},
		[]string{"foo=Accept: application/json", "bar=Authorization: Basic xxxxx"},
	)
	require.NoError(t, err)
	assert.EqualValues(t, &gomplate.Config{
		Templates: map[string]gomplate.DataSource{
			"foo": {
				URL:    mustURL("http://example.com"),
				Header: http.Header{"Accept": {"application/json"}},
			},
			"file.tmpl": {URL: mustURL("file.tmpl")},
			"tmpldir/":  {URL: mustURL("tmpldir/")},
		},
		ExtraHeaders: map[string]http.Header{
			"bar": {"Authorization": {"Basic xxxxx"}},
		},
	}, cfg)
}

func TestParsePluginFlags(t *testing.T) {
	t.Parallel()
	cfg := &gomplate.Config{}
	err := ParsePluginFlags(cfg, nil)
	require.NoError(t, err)

	cfg = &gomplate.Config{}
	err = ParsePluginFlags(cfg, []string{"foo=bar"})
	require.NoError(t, err)
	assert.EqualValues(t, &gomplate.Config{Plugins: map[string]gomplate.PluginConfig{"foo": {Cmd: "bar"}}}, cfg)
}

func TestParseDatasourceArgNoAlias(t *testing.T) {
	alias, ds, err := parseDatasourceArg("foo.json")
	require.NoError(t, err)
	assert.Equal(t, "foo", alias)
	assert.Empty(t, ds.URL.Scheme)

	_, _, err = parseDatasourceArg("../foo.json")
	require.Error(t, err)

	_, _, err = parseDatasourceArg("ftp://example.com/foo.yml")
	require.Error(t, err)
}

func TestParseDatasourceArgWithAlias(t *testing.T) {
	alias, ds, err := parseDatasourceArg("data=foo.json")
	require.NoError(t, err)
	assert.Equal(t, "data", alias)
	assert.EqualValues(t, &url.URL{Path: "foo.json"}, ds.URL)

	alias, ds, err = parseDatasourceArg("data=/otherdir/foo.json")
	require.NoError(t, err)
	assert.Equal(t, "data", alias)
	assert.Equal(t, "file", ds.URL.Scheme)
	assert.True(t, ds.URL.IsAbs())
	assert.Equal(t, "/otherdir/foo.json", ds.URL.Path)

	if runtime.GOOS == "windows" {
		alias, ds, err = parseDatasourceArg("data=foo.json")
		require.NoError(t, err)
		assert.Equal(t, "data", alias)
		assert.EqualValues(t, &url.URL{Path: "foo.json"}, ds.URL)

		alias, ds, err = parseDatasourceArg(`data=\otherdir\foo.json`)
		require.NoError(t, err)
		assert.Equal(t, "data", alias)
		assert.EqualValues(t, &url.URL{Scheme: "file", Path: "/otherdir/foo.json"}, ds.URL)

		alias, ds, err = parseDatasourceArg("data=C:\\windowsdir\\foo.json")
		require.NoError(t, err)
		assert.Equal(t, "data", alias)
		assert.EqualValues(t, &url.URL{Scheme: "file", Path: "C:/windowsdir/foo.json"}, ds.URL)

		alias, ds, err = parseDatasourceArg("data=\\\\somehost\\share\\foo.json")
		require.NoError(t, err)
		assert.Equal(t, "data", alias)
		assert.EqualValues(t, &url.URL{Scheme: "file", Host: "somehost", Path: "/share/foo.json"}, ds.URL)
	}

	alias, ds, err = parseDatasourceArg("data=sftp://example.com/blahblah/foo.json")
	require.NoError(t, err)
	assert.Equal(t, "data", alias)
	assert.EqualValues(t, &url.URL{Scheme: "sftp", Host: "example.com", Path: "/blahblah/foo.json"}, ds.URL)

	alias, ds, err = parseDatasourceArg("merged=merge:./foo.yaml|http://example.com/bar.json%3Ffoo=bar")
	require.NoError(t, err)
	assert.Equal(t, "merged", alias)
	assert.EqualValues(t, &url.URL{Scheme: "merge", Opaque: "./foo.yaml|http://example.com/bar.json%3Ffoo=bar"}, ds.URL)
}

func TestParseHeaderArgs(t *testing.T) {
	args := []string{
		"foo=Accept: application/json",
		"bar=Authorization: Bearer supersecret",
	}
	expected := map[string]http.Header{
		"foo": {
			"Accept": {"application/json"},
		},
		"bar": {
			"Authorization": {"Bearer supersecret"},
		},
	}
	parsed, err := parseHeaderArgs(args)
	require.NoError(t, err)
	assert.Equal(t, expected, parsed)

	_, err = parseHeaderArgs([]string{"foo"})
	require.Error(t, err)

	_, err = parseHeaderArgs([]string{"foo=bar"})
	require.Error(t, err)

	args = []string{
		"foo=Accept: application/json",
		"foo=Foo: bar",
		"foo=foo: baz",
		"foo=fOO: qux",
		"bar=Authorization: Bearer  supersecret",
	}
	expected = map[string]http.Header{
		"foo": {
			"Accept": {"application/json"},
			"Foo":    {"bar", "baz", "qux"},
		},
		"bar": {
			"Authorization": {"Bearer  supersecret"},
		},
	}
	parsed, err = parseHeaderArgs(args)
	require.NoError(t, err)
	assert.Equal(t, expected, parsed)
}
