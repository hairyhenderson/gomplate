package config

import (
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/hairyhenderson/gomplate/v4/internal/iohelpers"
	"github.com/hairyhenderson/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseConfigFile(t *testing.T) {
	t.Parallel()
	in := "in: hello world\n"
	expected := &Config{
		Input: "hello world",
	}
	cf, err := Parse(strings.NewReader(in))
	require.NoError(t, err)
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

plugins:
  foo:
    cmd: echo
    pipe: true

templates:
  foo:
    url: file:///tmp/foo.t

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
		OutMode: "644",
		Plugins: map[string]PluginConfig{
			"foo": {Cmd: "echo", Pipe: true},
		},
		Templates:     map[string]DataSource{"foo": {URL: mustURL("file:///tmp/foo.t")}},
		PluginTimeout: 2 * time.Second,
	}

	cf, err = Parse(strings.NewReader(in))
	require.NoError(t, err)
	assert.EqualValues(t, expected, cf)
}

func mustURL(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}

	return u
}

func TestValidate(t *testing.T) {
	t.Parallel()
	require.NoError(t, validateConfig(""))

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

	require.NoError(t, validateConfig(`execPipe: true
postExec: [echo, foo]
`))

	assert.Error(t, validateConfig(`execPipe: true
outputFiles: [foo]
postExec: [echo]
`))

	require.NoError(t, validateConfig(`execPipe: true
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
		Plugins: map[string]PluginConfig{
			"sleep": {Cmd: "echo"},
		},
		PluginTimeout: 500 * time.Microsecond,
	}
	other = &Config{
		InputFiles:  []string{"-"},
		OutputFiles: []string{"-"},
		Plugins: map[string]PluginConfig{
			"sleep": {Cmd: "sleep.sh"},
		},
	}
	expected = &Config{
		Input:       "hello world",
		OutputFiles: []string{"-"},
		Plugins: map[string]PluginConfig{
			"sleep": {Cmd: "sleep.sh"},
		},
		PluginTimeout: 500 * time.Microsecond,
	}

	assert.EqualValues(t, expected, cfg.MergeFrom(other))

	cfg = &Config{
		Input:   "hello world",
		OutMode: "644",
	}
	other = &Config{
		OutputFiles: []string{"out.txt"},
		Context: map[string]DataSource{
			"foo": {
				URL: mustURL("https://example.com/foo.yaml"),
				Header: http.Header{
					"Accept": {"application/json"},
				},
			},
			"bar": {URL: mustURL("stdin:///")},
		},
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
	}
	expected = &Config{
		Input:       "hello world",
		OutputFiles: []string{"out.txt"},
		Context: map[string]DataSource{
			"foo": {
				URL: mustURL("https://example.com/foo.yaml"),
				Header: http.Header{
					"Accept": {"application/json"},
				},
			},
			"bar": {URL: mustURL("stdin:///")},
		},
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
		OutMode: "644",
	}

	assert.EqualValues(t, expected, cfg.MergeFrom(other))

	// test template merging & a few other things
	cfg = &Config{
		InputDir:    "indir/",
		ExcludeGlob: []string{"*.txt"},
		Templates: map[string]DataSource{
			"foo": {
				URL: mustURL("file:///foo.yaml"),
			},
			"bar": {
				URL:    mustURL("stdin:///"),
				Header: http.Header{"Accept": {"application/json"}},
			},
		},
	}
	other = &Config{
		ExcludeGlob: []string{"*.yaml"},
		OutputMap:   "${ .in }.out",
		OutMode:     "600",
		LDelim:      "${",
		RDelim:      "}",
		Templates: map[string]DataSource{
			"foo": {URL: mustURL("https://example.com/foo.yaml")},
			"baz": {URL: mustURL("vault:///baz")},
		},
	}
	expected = &Config{
		InputDir:    "indir/",
		ExcludeGlob: []string{"*.yaml"},
		OutputMap:   "${ .in }.out",
		OutMode:     "600",
		LDelim:      "${",
		RDelim:      "}",
		Templates: map[string]DataSource{
			"foo": {URL: mustURL("https://example.com/foo.yaml")},
			"bar": {
				URL:    mustURL("stdin:///"),
				Header: http.Header{"Accept": {"application/json"}},
			},
			"baz": {URL: mustURL("vault:///baz")},
		},
	}

	assert.EqualValues(t, expected, cfg.MergeFrom(other))
}

func TestParseDataSourceFlags(t *testing.T) {
	t.Parallel()
	cfg := &Config{}
	err := cfg.ParseDataSourceFlags(nil, nil, nil, nil)
	require.NoError(t, err)
	assert.EqualValues(t, &Config{}, cfg)

	cfg = &Config{}
	err = cfg.ParseDataSourceFlags([]string{"foo/bar/baz.json"}, nil, nil, nil)
	assert.Error(t, err)

	cfg = &Config{}
	err = cfg.ParseDataSourceFlags([]string{"baz=foo/bar/baz.json"}, nil, nil, nil)
	require.NoError(t, err)
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
		nil,
		[]string{"baz=Accept: application/json"})
	require.NoError(t, err)
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
		nil,
		[]string{
			"foo=Accept: application/json",
			"bar=Authorization: Basic xxxxx",
		},
	)
	require.NoError(t, err)
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

	cfg = &Config{}
	err = cfg.ParseDataSourceFlags(
		nil,
		nil,
		[]string{"foo=http://example.com", "file.tmpl", "tmpldir/"},
		[]string{"foo=Accept: application/json", "bar=Authorization: Basic xxxxx"},
	)
	require.NoError(t, err)
	assert.EqualValues(t, &Config{
		Templates: map[string]DataSource{
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
	cfg := &Config{}
	err := cfg.ParsePluginFlags(nil)
	require.NoError(t, err)

	cfg = &Config{}
	err = cfg.ParsePluginFlags([]string{"foo=bar"})
	require.NoError(t, err)
	assert.EqualValues(t, &Config{Plugins: map[string]PluginConfig{"foo": {Cmd: "bar"}}}, cfg)
}

func TestConfig_String(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		c := &Config{}
		c.ApplyDefaults()

		expected := `---
inputFiles: ['-']
outputFiles: ['-']
leftDelim: '{{'
rightDelim: '}}'
missingKey: error
pluginTimeout: 5s
`
		assert.Equal(t, expected, c.String())
	})

	t.Run("overridden values", func(t *testing.T) {
		c := &Config{
			LDelim:      "L",
			RDelim:      "R",
			Input:       "foo",
			OutputFiles: []string{"-"},
			Templates: map[string]DataSource{
				"foo": {URL: mustURL("https://www.example.com/foo.tmpl")},
				"bar": {URL: mustURL("file:///tmp/bar.t")},
			},
		}
		expected := `---
in: foo
outputFiles: ['-']
leftDelim: L
rightDelim: R
templates:
  foo:
    url: https://www.example.com/foo.tmpl
  bar:
    url: file:///tmp/bar.t
`
		assert.YAMLEq(t, expected, c.String())
	})

	t.Run("long input", func(t *testing.T) {
		c := &Config{
			LDelim:      "L",
			RDelim:      "R",
			Input:       "long input that should be truncated",
			OutputFiles: []string{"-"},
			Templates: map[string]DataSource{
				"foo": {URL: mustURL("https://www.example.com/foo.tmpl")},
				"bar": {URL: mustURL("file:///tmp/bar.t")},
			},
		}
		expected := `---
in: long inp...
outputFiles: ['-']
leftDelim: L
rightDelim: R
templates:
  foo:
    url: https://www.example.com/foo.tmpl
  bar:
    url: file:///tmp/bar.t
`
		assert.YAMLEq(t, expected, c.String())
	})

	t.Run("relative dirs", func(t *testing.T) {
		c := &Config{
			InputDir:  "in/",
			OutputDir: "out/",
		}
		expected := `---
inputDir: in/
outputDir: out/
`
		assert.YAMLEq(t, expected, c.String())
	})

	t.Run("outputmap", func(t *testing.T) {
		c := &Config{
			InputDir:  "in/",
			OutputMap: "{{ .in }}",
		}
		expected := `---
inputDir: in/
outputMap: '{{ .in }}'
`

		assert.YAMLEq(t, expected, c.String())
	})

	t.Run("pluginTimeout", func(t *testing.T) {
		c := &Config{
			PluginTimeout: 500 * time.Millisecond,
		}
		expected := `---
pluginTimeout: 500ms
`

		assert.YAMLEq(t, expected, c.String())
	})

	t.Run("plugins", func(t *testing.T) {
		c := &Config{
			Plugins: map[string]PluginConfig{
				"foo": {
					Cmd:     "bar",
					Timeout: 1 * time.Second,
					Pipe:    true,
				},
			},
		}
		expected := `---
plugins:
  foo:
    cmd: bar
    timeout: 1s
    pipe: true
`

		assert.YAMLEq(t, expected, c.String())
	})
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
	require.NoError(t, err)
	assert.Equal(t, iohelpers.NormalizeFileMode(0), m)
	assert.False(t, o)

	c = &Config{OutMode: "755"}
	m, o, err = c.GetMode()
	require.NoError(t, err)
	assert.Equal(t, iohelpers.NormalizeFileMode(0o755), m)
	assert.True(t, o)

	c = &Config{OutMode: "0755"}
	m, o, err = c.GetMode()
	require.NoError(t, err)
	assert.Equal(t, iohelpers.NormalizeFileMode(0o755), m)
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
	require.NoError(t, err)
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
	require.NoError(t, err)
	assert.Equal(t, expected, parsed)
}

func TestParseDatasourceArgNoAlias(t *testing.T) {
	alias, ds, err := parseDatasourceArg("foo.json")
	require.NoError(t, err)
	assert.Equal(t, "foo", alias)
	assert.Empty(t, ds.URL.Scheme)

	_, _, err = parseDatasourceArg("../foo.json")
	assert.Error(t, err)

	_, _, err = parseDatasourceArg("ftp://example.com/foo.yml")
	assert.Error(t, err)
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

func TestPluginConfig_UnmarshalYAML(t *testing.T) {
	in := `foo`
	out := PluginConfig{}
	err := yaml.Unmarshal([]byte(in), &out)
	require.NoError(t, err)
	assert.EqualValues(t, PluginConfig{Cmd: "foo"}, out)

	in = `[foo, bar]`
	out = PluginConfig{}
	err = yaml.Unmarshal([]byte(in), &out)
	assert.Error(t, err)

	in = `cmd: foo`
	out = PluginConfig{}
	err = yaml.Unmarshal([]byte(in), &out)
	require.NoError(t, err)
	assert.EqualValues(t, PluginConfig{Cmd: "foo"}, out)

	in = `cmd: foo
timeout: 10ms
pipe: true
`
	out = PluginConfig{}
	err = yaml.Unmarshal([]byte(in), &out)
	require.NoError(t, err)
	assert.EqualValues(t, PluginConfig{
		Cmd:     "foo",
		Timeout: time.Duration(10) * time.Millisecond,
		Pipe:    true,
	}, out)
}
