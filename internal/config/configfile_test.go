package config

import (
	"net/http"
	"net/url"
	"os"
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
		DataSources: map[string]DSConfig{
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
		Context: map[string]DSConfig{
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
		DataSources: map[string]DSConfig{
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
		Context: map[string]DSConfig{
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
		DataSources: map[string]DSConfig{
			"data": {
				Header: http.Header{
					"Accept": {"foo/bar"},
				},
			},
		},
		Context: map[string]DSConfig{
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
		DataSources: map[string]DSConfig{
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
		Context: map[string]DSConfig{
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
		DataSources: DSources{
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
		DataSources: DSources{
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
		DataSources: DSources{
			"baz": {URL: mustURL("foo/bar/baz.json")},
		},
		Context: DSources{
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
	assert.Equal(t, os.FileMode(0), m)
	assert.False(t, o)

	c = &Config{OutMode: "755"}
	m, o, err = c.GetMode()
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0755), m)
	assert.True(t, o)

	c = &Config{OutMode: "0755"}
	m, o, err = c.GetMode()
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0755), m)
	assert.True(t, o)

	c = &Config{OutMode: "foo"}
	_, _, err = c.GetMode()
	assert.Error(t, err)
}
