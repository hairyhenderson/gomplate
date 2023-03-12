package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hairyhenderson/gomplate/v4/internal/config"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestReadConfigFile(t *testing.T) {
	fs = afero.NewMemMapFs()
	defer func() { fs = afero.NewOsFs() }()
	cmd := &cobra.Command{}

	_, err := readConfigFile(cmd)
	assert.NoError(t, err)

	cmd.Flags().String("config", defaultConfigFile, "foo")

	_, err = readConfigFile(cmd)
	assert.NoError(t, err)

	cmd.ParseFlags([]string{"--config", "config.file"})

	_, err = readConfigFile(cmd)
	assert.Error(t, err)

	cmd = &cobra.Command{}
	cmd.Flags().String("config", defaultConfigFile, "foo")

	f, err := fs.Create(defaultConfigFile)
	assert.NoError(t, err)
	f.WriteString("")

	cfg, err := readConfigFile(cmd)
	assert.NoError(t, err)
	assert.EqualValues(t, &config.Config{}, cfg)

	cmd.ParseFlags([]string{"--config", "config.yaml"})

	f, err = fs.Create("config.yaml")
	assert.NoError(t, err)
	f.WriteString("in: hello world\n")

	cfg, err = readConfigFile(cmd)
	assert.NoError(t, err)
	assert.EqualValues(t, &config.Config{Input: "hello world"}, cfg)

	f.WriteString("in: ")

	_, err = readConfigFile(cmd)
	assert.Error(t, err)
}

func TestLoadConfig(t *testing.T) {
	fs = afero.NewMemMapFs()
	defer func() { fs = afero.NewOsFs() }()

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

	out, err := loadConfig(cmd, cmd.Flags().Args())
	expected := &config.Config{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
	}
	assert.NoError(t, err)
	assert.EqualValues(t, expected, out)

	cmd.ParseFlags([]string{"--in", "foo"})
	out, err = loadConfig(cmd, cmd.Flags().Args())
	expected = &config.Config{
		Input:  "foo",
		Stdin:  stdin,
		Stdout: out.Stdout,
		Stderr: stderr,
	}
	assert.NoError(t, err)
	assert.EqualValues(t, expected, out)

	cmd.ParseFlags([]string{"--in", "foo", "--exec-pipe", "--", "tr", "[a-z]", "[A-Z]"})
	out, err = loadConfig(cmd, cmd.Flags().Args())
	expected = &config.Config{
		Input:         "foo",
		ExecPipe:      true,
		PostExec:      []string{"tr", "[a-z]", "[A-Z]"},
		PostExecInput: out.PostExecInput,
		Stdin:         stdin,
		Stdout:        out.Stdout,
		Stderr:        stderr,
	}
	assert.NoError(t, err)
	assert.EqualValues(t, expected, out)
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
	assert.NoError(t, err)
	assert.EqualValues(t, &config.Config{}, cfg)

	cmd.ParseFlags([]string{"--file", "in", "--", "echo", "foo"})

	cfg, err = cobraConfig(cmd, cmd.Flags().Args())
	assert.NoError(t, err)
	assert.EqualValues(t, &config.Config{
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

	cf, req := pickConfigFile(cmd)
	assert.False(t, req)
	assert.Equal(t, defaultConfigFile, cf)

	os.Setenv("GOMPLATE_CONFIG", "foo.yaml")
	defer os.Unsetenv("GOMPLATE_CONFIG")
	cf, req = pickConfigFile(cmd)
	assert.True(t, req)
	assert.Equal(t, "foo.yaml", cf)

	cmd.ParseFlags([]string{"--config", "config.file"})
	cf, req = pickConfigFile(cmd)
	assert.True(t, req)
	assert.Equal(t, "config.file", cf)

	os.Setenv("GOMPLATE_CONFIG", "ignored.yaml")
	cf, req = pickConfigFile(cmd)
	assert.True(t, req)
	assert.Equal(t, "config.file", cf)
}

func TestApplyEnvVars(t *testing.T) {
	os.Setenv("GOMPLATE_PLUGIN_TIMEOUT", "bogus")
	_, err := applyEnvVars(context.Background(), &config.Config{})
	os.Unsetenv("GOMPLATE_PLUGIN_TIMEOUT")
	assert.Error(t, err)

	data := []struct {
		input, expected *config.Config
		env             string
		value           string
	}{
		{
			&config.Config{PluginTimeout: 2 * time.Second},
			&config.Config{PluginTimeout: 2 * time.Second},
			"GOMPLATE_PLUGIN_TIMEOUT", "bogus",
		},
		{
			&config.Config{},
			&config.Config{PluginTimeout: 2 * time.Second},
			"GOMPLATE_PLUGIN_TIMEOUT", "2s",
		},
		{
			&config.Config{PluginTimeout: 100 * time.Millisecond},
			&config.Config{PluginTimeout: 100 * time.Millisecond},
			"GOMPLATE_PLUGIN_TIMEOUT", "2s",
		},
		{
			&config.Config{},
			&config.Config{SuppressEmpty: false},
			"GOMPLATE_SUPPRESS_EMPTY", "bogus",
		},
		{
			&config.Config{},
			&config.Config{SuppressEmpty: true},
			"GOMPLATE_SUPPRESS_EMPTY", "true",
		},
		{
			&config.Config{SuppressEmpty: true},
			&config.Config{SuppressEmpty: true},
			"GOMPLATE_SUPPRESS_EMPTY", "false",
		},
		{
			&config.Config{},
			&config.Config{Experimental: false},
			"GOMPLATE_EXPERIMENTAL", "bogus",
		},
		{
			&config.Config{},
			&config.Config{Experimental: true},
			"GOMPLATE_EXPERIMENTAL", "true",
		},
		{
			&config.Config{Experimental: true},
			&config.Config{Experimental: true},
			"GOMPLATE_EXPERIMENTAL", "false",
		},
		{
			&config.Config{},
			&config.Config{LDelim: "--"},
			"GOMPLATE_LEFT_DELIM", "--",
		},
		{
			&config.Config{LDelim: "{{"},
			&config.Config{LDelim: "{{"},
			"GOMPLATE_LEFT_DELIM", "--",
		},
		{
			&config.Config{},
			&config.Config{RDelim: ")>"},
			"GOMPLATE_RIGHT_DELIM", ")>",
		},
		{
			&config.Config{RDelim: "}}"},
			&config.Config{RDelim: "}}"},
			"GOMPLATE_RIGHT_DELIM", ")>",
		},
		{
			&config.Config{RDelim: "}}"},
			&config.Config{RDelim: "}}"},
			"GOMPLATE_RIGHT_DELIM", "",
		},
	}

	for i, d := range data {
		d := d
		t.Run(fmt.Sprintf("applyEnvVars_%s_%s/%d", d.env, d.value, i), func(t *testing.T) {
			os.Setenv(d.env, d.value)

			actual, err := applyEnvVars(context.Background(), d.input)
			os.Unsetenv(d.env)
			assert.NoError(t, err)
			assert.EqualValues(t, d.expected, actual)
		})
	}
}
