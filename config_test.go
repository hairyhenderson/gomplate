package gomplate

import (
	"net/http"
	"net/url"
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
	assert.Equal(t, expected, cf)
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

	require.Error(t, validateConfig(`in: foo
inputFiles: [bar]
`))
	require.Error(t, validateConfig(`inputDir: foo
inputFiles: [bar]
`))
	require.Error(t, validateConfig(`inputDir: foo
in: bar
`))

	require.Error(t, validateConfig(`outputDir: foo
outputFiles: [bar]
`))

	require.Error(t, validateConfig(`in: foo
outputFiles: [bar, baz]
`))

	require.Error(t, validateConfig(`inputFiles: [foo]
outputFiles: [bar, baz]
`))

	require.Error(t, validateConfig(`outputDir: foo
outputFiles: [bar]
`))

	require.Error(t, validateConfig(`outputDir: foo
`))

	require.Error(t, validateConfig(`outputMap: foo
`))

	require.Error(t, validateConfig(`outputMap: foo
outputFiles: [bar]
`))

	require.Error(t, validateConfig(`inputDir: foo
outputDir: bar
outputMap: bar
`))

	require.Error(t, validateConfig(`execPipe: true
`))
	require.Error(t, validateConfig(`execPipe: true
postExec: ""
`))

	require.NoError(t, validateConfig(`execPipe: true
postExec: [echo, foo]
`))

	require.Error(t, validateConfig(`execPipe: true
outputFiles: [foo]
postExec: [echo]
`))

	require.NoError(t, validateConfig(`execPipe: true
inputFiles: ['-']
postExec: [echo]
`))

	require.Error(t, validateConfig(`inputDir: foo
execPipe: true
outputDir: foo
postExec: [echo]
`))

	require.Error(t, validateConfig(`inputDir: foo
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
	err = cfg.validate()
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

	assert.Equal(t, expected, cfg.MergeFrom(other))

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

	assert.Equal(t, expected, cfg.MergeFrom(other))

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

	assert.Equal(t, expected, cfg.MergeFrom(other))

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

	assert.Equal(t, expected, cfg.MergeFrom(other))

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

	assert.Equal(t, expected, cfg.MergeFrom(other))

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

	assert.Equal(t, expected, cfg.MergeFrom(other))

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

	assert.Equal(t, expected, cfg.MergeFrom(other))
}

func TestConfig_String(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		c := &Config{}
		c.applyDefaults()

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
    header: {}
  bar:
    url: file:///tmp/bar.t
    header: {}
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
    header: {}
  bar:
    url: file:///tmp/bar.t
    header: {}
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

	cfg.applyDefaults()
	assert.Equal(t, []string{"-"}, cfg.InputFiles)
	assert.Equal(t, []string{"-"}, cfg.OutputFiles)
	assert.Empty(t, cfg.OutputDir)
	assert.Equal(t, "{{", cfg.LDelim)
	assert.Equal(t, "}}", cfg.RDelim)

	cfg = &Config{
		InputDir: "in",
	}

	cfg.applyDefaults()
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

	cfg.applyDefaults()
	assert.Empty(t, cfg.InputFiles)
	assert.Equal(t, []string{"-"}, cfg.OutputFiles)
	assert.Empty(t, cfg.OutputDir)
	assert.Equal(t, "<", cfg.LDelim)
	assert.Equal(t, ">", cfg.RDelim)

	cfg = &Config{
		Input:    "foo",
		ExecPipe: true,
	}

	cfg.applyDefaults()
	assert.Empty(t, cfg.InputFiles)
	assert.Equal(t, []string{"-"}, cfg.OutputFiles)
	assert.Empty(t, cfg.OutputDir)
	assert.True(t, cfg.ExecPipe)

	cfg = &Config{
		InputDir:  "foo",
		OutputMap: "bar",
	}

	cfg.applyDefaults()
	assert.Empty(t, cfg.InputFiles)
	assert.Empty(t, cfg.Input)
	assert.Empty(t, cfg.OutputFiles)
	assert.Empty(t, cfg.OutputDir)
	assert.False(t, cfg.ExecPipe)
	assert.Equal(t, "bar", cfg.OutputMap)
}

func TestGetMode(t *testing.T) {
	c := &Config{}
	m, o, err := c.getMode()
	require.NoError(t, err)
	assert.Equal(t, iohelpers.NormalizeFileMode(0), m)
	assert.False(t, o)

	c = &Config{OutMode: "755"}
	m, o, err = c.getMode()
	require.NoError(t, err)
	assert.Equal(t, iohelpers.NormalizeFileMode(0o755), m)
	assert.True(t, o)

	c = &Config{OutMode: "0755"}
	m, o, err = c.getMode()
	require.NoError(t, err)
	assert.Equal(t, iohelpers.NormalizeFileMode(0o755), m)
	assert.True(t, o)

	c = &Config{OutMode: "foo"}
	_, _, err = c.getMode()
	require.Error(t, err)
}

func TestPluginConfig_UnmarshalYAML(t *testing.T) {
	in := `foo`
	out := PluginConfig{}
	err := yaml.Unmarshal([]byte(in), &out)
	require.NoError(t, err)
	assert.Equal(t, PluginConfig{Cmd: "foo"}, out)

	in = `[foo, bar]`
	out = PluginConfig{}
	err = yaml.Unmarshal([]byte(in), &out)
	require.Error(t, err)

	in = `cmd: foo`
	out = PluginConfig{}
	err = yaml.Unmarshal([]byte(in), &out)
	require.NoError(t, err)
	assert.Equal(t, PluginConfig{Cmd: "foo"}, out)

	in = `cmd: foo
timeout: 10ms
pipe: true
`
	out = PluginConfig{}
	err = yaml.Unmarshal([]byte(in), &out)
	require.NoError(t, err)
	assert.Equal(t, PluginConfig{
		Cmd:     "foo",
		Timeout: time.Duration(10) * time.Millisecond,
		Pipe:    true,
	}, out)
}
