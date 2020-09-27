package config

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseConfigFile(t *testing.T) {
	t.Parallel()
	in := "in: hello world\n"
	expected := &Config{
		Input: "hello world",
	}
	cf, err := Parse(strings.NewReader(in))
	assert.NoError(t, err)
	assert.Equal(t, expected, cf)

	in = `in: hello world
outputFiles: [out.txt]
chmod: 644

datasources:
  data:
    url: file:///data.json
  moredata:
    url: https://example.com/more.json
    header:
      Authorization: ["Bearer abcd1234"]

context:
  .:
    url: file:///data.json

pluginTimeout: 2s
`
	expected = &Config{
		Input:       "hello world",
		OutputFiles: []string{"out.txt"},
		DataSources: map[string]DataSource{
			"data": {
				URL: mustURL("file:///data.json"),
			},
			"moredata": {
				URL: mustURL("https://example.com/more.json"),
				Header: map[string][]string{
					"Authorization": {"Bearer abcd1234"},
				},
			},
		},
		Context: map[string]DataSource{
			".": {
				URL: mustURL("file:///data.json"),
			},
		},
		OutMode:       "644",
		PluginTimeout: 2 * time.Second,
	}

	cf, err = Parse(strings.NewReader(in))
	assert.NoError(t, err)
	assert.EqualValues(t, expected, cf)
}

func mustURL(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	// handle the case where it's a relative URL - just like in parseSourceURL.
	if !u.IsAbs() {
		u, err = absFileURL(s)
		if err != nil {
			panic(err)
		}
	}
	return u
}

func TestValidate(t *testing.T) {
	t.Parallel()
	assert.NoError(t, validateConfig(""))

	assert.Error(t, validateConfig(`in: foo
inputFiles: [bar]
`))
	assert.Error(t, validateConfig(`inputDir: foo
inputFiles: [bar]
`))
	assert.Error(t, validateConfig(`inputDir: foo
in: bar
`))

	assert.Error(t, validateConfig(`outputDir: foo
outputFiles: [bar]
`))

	assert.Error(t, validateConfig(`in: foo
outputFiles: [bar, baz]
`))

	assert.Error(t, validateConfig(`inputFiles: [foo]
outputFiles: [bar, baz]
`))

	assert.Error(t, validateConfig(`outputDir: foo
outputFiles: [bar]
`))

	assert.Error(t, validateConfig(`outputDir: foo
`))

	assert.Error(t, validateConfig(`outputMap: foo
`))

	assert.Error(t, validateConfig(`outputMap: foo
outputFiles: [bar]
`))

	assert.Error(t, validateConfig(`inputDir: foo
outputDir: bar
outputMap: bar
`))

	assert.Error(t, validateConfig(`execPipe: true
`))
	assert.Error(t, validateConfig(`execPipe: true
postExec: ""
`))

	assert.NoError(t, validateConfig(`execPipe: true
postExec: [echo, foo]
`))

	assert.Error(t, validateConfig(`execPipe: true
outputFiles: [foo]
postExec: [echo]
`))

	assert.NoError(t, validateConfig(`execPipe: true
inputFiles: ['-']
postExec: [echo]
`))

	assert.Error(t, validateConfig(`inputDir: foo
execPipe: true
outputDir: foo
postExec: [echo]
`))

	assert.Error(t, validateConfig(`inputDir: foo
execPipe: true
outputMap: foo
postExec: [echo]
`))
}

func validateConfig(c string) error {
	in := strings.NewReader(c)
	cfg, err := Parse(in)
	if err != nil {
		return err
	}
	err = cfg.Validate()
	return err
}

func TestMergeFrom(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		Input: "hello world",
		DataSources: map[string]DataSource{
			"data": {
				URL: mustURL("file:///data.json"),
			},
			"moredata": {
				URL: mustURL("https://example.com/more.json"),
				Header: http.Header{
					"Authorization": {"Bearer abcd1234"},
				},
			},
		},
		Context: map[string]DataSource{
			"foo": {
				URL: mustURL("https://example.com/foo.yaml"),
				Header: http.Header{
					"Accept": {"application/yaml"},
				},
			},
		},
		OutMode: "644",
	}
	other := &Config{
		OutputFiles: []string{"out.txt"},
		DataSources: map[string]DataSource{
			"data": {
				Header: http.Header{
					"Accept": {"foo/bar"},
				},
			},
		},
		Context: map[string]DataSource{
			"foo": {
				Header: http.Header{
					"Accept": {"application/json"},
				},
			},
			"bar": {URL: mustURL("stdin:///")},
		},
	}
	expected := &Config{
		Input:       "hello world",
		OutputFiles: []string{"out.txt"},
		DataSources: map[string]DataSource{
			"data": {
				URL: mustURL("file:///data.json"),
				Header: http.Header{
					"Accept": {"foo/bar"},
				},
			},
			"moredata": {
				URL: mustURL("https://example.com/more.json"),
				Header: http.Header{
					"Authorization": {"Bearer abcd1234"},
				},
			},
		},
		Context: map[string]DataSource{
			"foo": {
				URL: mustURL("https://example.com/foo.yaml"),
				Header: http.Header{
					"Accept": {"application/json"},
				},
			},
			"bar": {URL: mustURL("stdin:///")},
		},
		OutMode: "644",
	}

	assert.EqualValues(t, expected, cfg.MergeFrom(other))

	cfg = &Config{
		Input: "hello world",
	}
	other = &Config{
		InputFiles:  []string{"in.tmpl", "in2.tmpl"},
		OutputFiles: []string{"out", "out2"},
	}
	expected = &Config{
		InputFiles:  []string{"in.tmpl", "in2.tmpl"},
		OutputFiles: []string{"out", "out2"},
	}

	assert.EqualValues(t, expected, cfg.MergeFrom(other))

	cfg = &Config{
		Input:       "hello world",
		OutputFiles: []string{"out", "out2"},
	}
	other = &Config{
		InputDir:  "in/",
		OutputDir: "out/",
	}
	expected = &Config{
		InputDir:  "in/",
		OutputDir: "out/",
	}

	assert.EqualValues(t, expected, cfg.MergeFrom(other))

	cfg = &Config{
		Input:       "hello world",
		OutputFiles: []string{"out"},
	}
	other = &Config{
		Input:    "hi",
		ExecPipe: true,
		PostExec: []string{"cat"},
	}
	expected = &Config{
		Input:    "hi",
		ExecPipe: true,
		PostExec: []string{"cat"},
	}

	assert.EqualValues(t, expected, cfg.MergeFrom(other))

	cfg = &Config{
		Input:       "hello world",
		OutputFiles: []string{"-"},
		Plugins: map[string]string{
			"sleep": "echo",
		},
		PluginTimeout: 500 * time.Microsecond,
	}
	other = &Config{
		InputFiles:  []string{"-"},
		OutputFiles: []string{"-"},
		Plugins: map[string]string{
			"sleep": "sleep.sh",
		},
	}
	expected = &Config{
		Input:       "hello world",
		OutputFiles: []string{"-"},
		Plugins: map[string]string{
			"sleep": "sleep.sh",
		},
		PluginTimeout: 500 * time.Microsecond,
	}

	assert.EqualValues(t, expected, cfg.MergeFrom(other))
}

func TestParseDataSourceFlags(t *testing.T) {
	t.Parallel()
	cfg := &Config{}
	err := cfg.ParseDataSourceFlags(nil, nil, nil)
	assert.NoError(t, err)
	assert.EqualValues(t, &Config{}, cfg)

	cfg = &Config{}
	err = cfg.ParseDataSourceFlags([]string{"foo/bar/baz.json"}, nil, nil)
	assert.Error(t, err)

	cfg = &Config{}
	err = cfg.ParseDataSourceFlags([]string{"baz=foo/bar/baz.json"}, nil, nil)
	assert.NoError(t, err)
	expected := &Config{
		DataSources: map[string]DataSource{
			"baz": {URL: mustURL("foo/bar/baz.json")},
		},
	}
	assert.EqualValues(t, expected, cfg, "expected: %+v\nactual: %+v\n", expected, cfg)

	cfg = &Config{}
	err = cfg.ParseDataSourceFlags(
		[]string{"baz=foo/bar/baz.json"},
		nil,
		[]string{"baz=Accept: application/json"})
	assert.NoError(t, err)
	assert.EqualValues(t, &Config{
		DataSources: map[string]DataSource{
			"baz": {
				URL: mustURL("foo/bar/baz.json"),
				Header: http.Header{
					"Accept": {"application/json"},
				},
			},
		},
	}, cfg)

	cfg = &Config{}
	err = cfg.ParseDataSourceFlags(
		[]string{"baz=foo/bar/baz.json"},
		[]string{"foo=http://example.com"},
		[]string{"foo=Accept: application/json",
			"bar=Authorization: Basic xxxxx"})
	assert.NoError(t, err)
	assert.EqualValues(t, &Config{
		DataSources: map[string]DataSource{
			"baz": {URL: mustURL("foo/bar/baz.json")},
		},
		Context: map[string]DataSource{
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
}

func TestParsePluginFlags(t *testing.T) {
	t.Parallel()
	cfg := &Config{}
	err := cfg.ParsePluginFlags(nil)
	assert.NoError(t, err)

	cfg = &Config{}
	err = cfg.ParsePluginFlags([]string{"foo=bar"})
	assert.NoError(t, err)
	assert.EqualValues(t, &Config{Plugins: map[string]string{"foo": "bar"}}, cfg)
}

func TestConfigString(t *testing.T) {
	c := &Config{}
	c.ApplyDefaults()

	expected := `---
inputFiles: ['-']
outputFiles: ['-']
leftDelim: '{{'
rightDelim: '}}'
pluginTimeout: 5s
`
	assert.Equal(t, expected, c.String())

	c = &Config{
		LDelim:      "L",
		RDelim:      "R",
		Input:       "foo",
		OutputFiles: []string{"-"},
		Templates:   []string{"foo=foo.t", "bar=bar.t"},
	}
	expected = `---
in: foo
outputFiles: ['-']
leftDelim: L
rightDelim: R
templates:
  - foo=foo.t
  - bar=bar.t
`
	assert.Equal(t, expected, c.String())

	c = &Config{
		LDelim:      "L",
		RDelim:      "R",
		Input:       "long input that should be truncated",
		OutputFiles: []string{"-"},
		Templates:   []string{"foo=foo.t", "bar=bar.t"},
	}
	expected = `---
in: long inp...
outputFiles: ['-']
leftDelim: L
rightDelim: R
templates:
  - foo=foo.t
  - bar=bar.t
`
	assert.Equal(t, expected, c.String())

	c = &Config{
		InputDir:  "in/",
		OutputDir: "out/",
	}
	expected = `---
inputDir: in/
outputDir: out/
`

	assert.Equal(t, expected, c.String())

	c = &Config{
		InputDir:  "in/",
		OutputMap: "{{ .in }}",
	}
	expected = `---
inputDir: in/
outputMap: '{{ .in }}'
`

	assert.Equal(t, expected, c.String())

	c = &Config{
		PluginTimeout: 500 * time.Millisecond,
	}
	expected = `---
pluginTimeout: 500ms
`

	assert.Equal(t, expected, c.String())
}

func TestApplyDefaults(t *testing.T) {
	t.Parallel()
	cfg := &Config{}

	cfg.ApplyDefaults()
	assert.EqualValues(t, []string{"-"}, cfg.InputFiles)
	assert.EqualValues(t, []string{"-"}, cfg.OutputFiles)
	assert.Empty(t, cfg.OutputDir)
	assert.Equal(t, "{{", cfg.LDelim)
	assert.Equal(t, "}}", cfg.RDelim)

	cfg = &Config{
		InputDir: "in",
	}

	cfg.ApplyDefaults()
	assert.Empty(t, cfg.InputFiles)
	assert.Empty(t, cfg.OutputFiles)
	assert.Equal(t, ".", cfg.OutputDir)
	assert.Equal(t, "{{", cfg.LDelim)
	assert.Equal(t, "}}", cfg.RDelim)

	cfg = &Config{
		Input:  "foo",
		LDelim: "<",
		RDelim: ">",
	}

	cfg.ApplyDefaults()
	assert.Empty(t, cfg.InputFiles)
	assert.EqualValues(t, []string{"-"}, cfg.OutputFiles)
	assert.Empty(t, cfg.OutputDir)
	assert.Equal(t, "<", cfg.LDelim)
	assert.Equal(t, ">", cfg.RDelim)

	cfg = &Config{
		Input:    "foo",
		ExecPipe: true,
	}

	cfg.ApplyDefaults()
	assert.Empty(t, cfg.InputFiles)
	assert.EqualValues(t, []string{"-"}, cfg.OutputFiles)
	assert.Empty(t, cfg.OutputDir)
	assert.True(t, cfg.ExecPipe)

	cfg = &Config{
		InputDir:  "foo",
		OutputMap: "bar",
	}

	cfg.ApplyDefaults()
	assert.Empty(t, cfg.InputFiles)
	assert.Empty(t, cfg.Input)
	assert.Empty(t, cfg.OutputFiles)
	assert.Empty(t, cfg.OutputDir)
	assert.False(t, cfg.ExecPipe)
	assert.Equal(t, "bar", cfg.OutputMap)
}

func TestGetMode(t *testing.T) {
	c := &Config{}
	m, o, err := c.GetMode()
	assert.NoError(t, err)
	assert.Equal(t, NormalizeFileMode(0), m)
	assert.False(t, o)

	c = &Config{OutMode: "755"}
	m, o, err = c.GetMode()
	assert.NoError(t, err)
	assert.Equal(t, NormalizeFileMode(0o755), m)
	assert.True(t, o)

	c = &Config{OutMode: "0755"}
	m, o, err = c.GetMode()
	assert.NoError(t, err)
	assert.Equal(t, NormalizeFileMode(0o755), m)
	assert.True(t, o)

	c = &Config{OutMode: "foo"}
	_, _, err = c.GetMode()
	assert.Error(t, err)
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
	assert.NoError(t, err)
	assert.Equal(t, expected, parsed)

	_, err = parseHeaderArgs([]string{"foo"})
	assert.Error(t, err)

	_, err = parseHeaderArgs([]string{"foo=bar"})
	assert.Error(t, err)

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
	assert.NoError(t, err)
	assert.Equal(t, expected, parsed)
}

func TestParseSourceURL(t *testing.T) {
	expected := &url.URL{
		Scheme:   "http",
		Host:     "example.com",
		Path:     "/foo.json",
		RawQuery: "bar",
	}
	u, err := ParseSourceURL("http://example.com/foo.json?bar")
	assert.NoError(t, err)
	assert.EqualValues(t, expected, u)

	expected = &url.URL{Scheme: "stdin"}
	u, err = ParseSourceURL("-")
	assert.NoError(t, err)
	assert.EqualValues(t, expected, u)

	wd, err := os.Getwd()
	assert.NoError(t, err)
	expected = &url.URL{
		Scheme: "file",
		Path:   path.Join(filepath.ToSlash(wd), "foo/bar.json"),
	}
	u, err = ParseSourceURL("./foo/bar.json")
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

func TestParseDatasourceArgNoAlias(t *testing.T) {
	key, ds, err := parseDatasourceArg("foo.json")
	assert.NoError(t, err)
	assert.Equal(t, "foo", key)
	assert.Equal(t, "file", ds.URL.Scheme)

	_, _, err = parseDatasourceArg("../foo.json")
	assert.Error(t, err)

	_, _, err = parseDatasourceArg("ftp://example.com/foo.yml")
	assert.Error(t, err)
}

func TestParseDatasourceArgWithAlias(t *testing.T) {
	key, ds, err := parseDatasourceArg("data=foo.json")
	assert.NoError(t, err)
	assert.Equal(t, "data", key)
	assert.Equal(t, "file", ds.URL.Scheme)
	assert.True(t, ds.URL.IsAbs())

	key, ds, err = parseDatasourceArg("data=/otherdir/foo.json")
	assert.NoError(t, err)
	assert.Equal(t, "data", key)
	assert.Equal(t, "file", ds.URL.Scheme)
	assert.True(t, ds.URL.IsAbs())
	assert.Equal(t, "/otherdir/foo.json", ds.URL.Path)

	if runtime.GOOS == "windows" {
		key, ds, err = parseDatasourceArg("data=foo.json")
		assert.NoError(t, err)
		assert.Equal(t, "data", key)
		assert.Equal(t, "file", ds.URL.Scheme)
		assert.True(t, ds.URL.IsAbs())
		assert.Equalf(t, byte(':'), ds.URL.Path[1], "Path was %s", ds.URL.Path)

		key, ds, err = parseDatasourceArg(`data=\otherdir\foo.json`)
		assert.NoError(t, err)
		assert.Equal(t, "data", key)
		assert.Equal(t, "file", ds.URL.Scheme)
		assert.True(t, ds.URL.IsAbs())
		assert.Equal(t, `/otherdir/foo.json`, ds.URL.Path)

		key, ds, err = parseDatasourceArg("data=C:\\windowsdir\\foo.json")
		assert.NoError(t, err)
		assert.Equal(t, "data", key)
		assert.Equal(t, "file", ds.URL.Scheme)
		assert.True(t, ds.URL.IsAbs())
		assert.Equal(t, "C:/windowsdir/foo.json", ds.URL.Path)

		key, ds, err = parseDatasourceArg("data=\\\\somehost\\share\\foo.json")
		assert.NoError(t, err)
		assert.Equal(t, "data", key)
		assert.Equal(t, "file", ds.URL.Scheme)
		assert.Equal(t, "somehost", ds.URL.Host)
		assert.True(t, ds.URL.IsAbs())
		assert.Equal(t, "/share/foo.json", ds.URL.Path)
	}

	key, ds, err = parseDatasourceArg("data=sftp://example.com/blahblah/foo.json")
	assert.NoError(t, err)
	assert.Equal(t, "data", key)
	assert.Equal(t, "sftp", ds.URL.Scheme)
	assert.True(t, ds.URL.IsAbs())
	assert.Equal(t, "/blahblah/foo.json", ds.URL.Path)

	key, ds, err = parseDatasourceArg("merged=merge:./foo.yaml|http://example.com/bar.json%3Ffoo=bar")
	assert.NoError(t, err)
	assert.Equal(t, "merged", key)
	assert.Equal(t, "merge", ds.URL.Scheme)
	assert.Equal(t, "./foo.yaml|http://example.com/bar.json%3Ffoo=bar", ds.URL.Opaque)
}

func TestContextWithConfig(t *testing.T) {
	ctx := context.Background()
	cfg := &Config{}
	ctx = ContextWithConfig(ctx, cfg)
	assert.Equal(t, cfg, ctx.Value(cfgContextKey))
}

func TestFromContext(t *testing.T) {
	cfg := &Config{}
	ctx := context.WithValue(context.Background(),
		cfgContextKey, cfg)
	assert.Equal(t, cfg, FromContext(ctx))

	ctx = context.Background()
	cfg = FromContext(ctx)
	// assert that the returned config looks like a default one
	assert.Equal(t, "{{", cfg.LDelim)
}
