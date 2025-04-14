package gomplate

import (
	"fmt"
	"io"
	"maps"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/hairyhenderson/gomplate/v4/internal/config"
	"github.com/hairyhenderson/gomplate/v4/internal/iohelpers"
	"github.com/hairyhenderson/yaml"
)

// Parse a config file
func Parse(in io.Reader) (*Config, error) {
	out := &Config{}
	dec := yaml.NewDecoder(in)
	err := dec.Decode(out)
	if err != nil && err != io.EOF {
		return out, fmt.Errorf("YAML decoding failed, syntax may be invalid: %w", err)
	}
	return out, nil
}

// Config models gomplate's configuration file and command-line options. It
// also contains some fields that can't be set in the config file.
type Config struct {
	// Stdin - override for stdin:// URLs or the '-' input file. Can't be set in
	// the config file.
	// Usually this should be left as default - this will be set at runtime.
	Stdin io.Reader `yaml:"-"`

	// Stdout - override for the '-' output file. Can't be set in the config
	// file.
	// Usually this should be left as default - this will be set at runtime.
	Stdout io.Writer `yaml:"-"`

	// Stderr - override for plugins to write to stderr. Can't be set in the
	// config file.
	// Usually this should be left as default - this will be set at runtime.
	Stderr io.Writer `yaml:"-"`

	// ExtraHeaders - Extra HTTP headers not attached to pre-defined datsources.
	// Potentially used by datasources defined in the template at runtime. Can't
	// currently be set in the config file.
	ExtraHeaders map[string]http.Header `yaml:"-"`

	DataSources map[string]DataSource   `yaml:"datasources,omitempty"`
	Context     map[string]DataSource   `yaml:"context,omitempty"`
	Templates   map[string]DataSource   `yaml:"templates,omitempty"`
	Plugins     map[string]PluginConfig `yaml:"plugins,omitempty"`

	Input                 string   `yaml:"in,omitempty"`
	InputDir              string   `yaml:"inputDir,omitempty"`
	InputFiles            []string `yaml:"inputFiles,omitempty,flow"`
	ExcludeGlob           []string `yaml:"excludes,omitempty"`
	ExcludeProcessingGlob []string `yaml:"excludeProcessing,omitempty"`

	OutputDir   string   `yaml:"outputDir,omitempty"`
	OutputMap   string   `yaml:"outputMap,omitempty"`
	OutputFiles []string `yaml:"outputFiles,omitempty,flow"`
	OutMode     string   `yaml:"chmod,omitempty"`

	LDelim string `yaml:"leftDelim,omitempty"`
	RDelim string `yaml:"rightDelim,omitempty"`

	MissingKey string `yaml:"missingKey,omitempty"`

	PostExec []string `yaml:"postExec,omitempty,flow"`

	PluginTimeout time.Duration `yaml:"pluginTimeout,omitempty"`

	ExecPipe     bool `yaml:"execPipe,omitempty"`
	Experimental bool `yaml:"experimental,omitempty"`
}

// TODO: remove when we remove the deprecated array format for templates
type rawConfig struct {
	DataSources map[string]DataSource   `yaml:"datasources,omitempty"`
	Context     map[string]DataSource   `yaml:"context,omitempty"`
	Templates   config.Templates        `yaml:"templates,omitempty"`
	Plugins     map[string]PluginConfig `yaml:"plugins,omitempty"`

	Input                 string   `yaml:"in,omitempty"`
	InputDir              string   `yaml:"inputDir,omitempty"`
	InputFiles            []string `yaml:"inputFiles,omitempty,flow"`
	ExcludeGlob           []string `yaml:"excludes,omitempty"`
	ExcludeProcessingGlob []string `yaml:"excludeProcessing,omitempty"`

	OutputDir   string   `yaml:"outputDir,omitempty"`
	OutputMap   string   `yaml:"outputMap,omitempty"`
	OutputFiles []string `yaml:"outputFiles,omitempty,flow"`
	OutMode     string   `yaml:"chmod,omitempty"`

	LDelim string `yaml:"leftDelim,omitempty"`
	RDelim string `yaml:"rightDelim,omitempty"`

	MissingKey string `yaml:"missingKey,omitempty"`

	PostExec []string `yaml:"postExec,omitempty,flow"`

	PluginTimeout time.Duration `yaml:"pluginTimeout,omitempty"`

	ExecPipe     bool `yaml:"execPipe,omitempty"`
	Experimental bool `yaml:"experimental,omitempty"`
}

// TODO: remove when we remove the deprecated array format for templates
//
// Deprecated: custom unmarshalling will be removed in the next version
func (c *Config) UnmarshalYAML(value *yaml.Node) error {
	r := rawConfig{}
	err := value.Decode(&r)
	if err != nil {
		return err
	}

	*c = Config{
		DataSources:           r.DataSources,
		Context:               r.Context,
		Templates:             r.Templates,
		Plugins:               r.Plugins,
		Input:                 r.Input,
		InputDir:              r.InputDir,
		InputFiles:            r.InputFiles,
		ExcludeGlob:           r.ExcludeGlob,
		ExcludeProcessingGlob: r.ExcludeProcessingGlob,
		OutputDir:             r.OutputDir,
		OutputMap:             r.OutputMap,
		OutputFiles:           r.OutputFiles,
		OutMode:               r.OutMode,
		LDelim:                r.LDelim,
		RDelim:                r.RDelim,
		MissingKey:            r.MissingKey,
		PostExec:              r.PostExec,
		PluginTimeout:         r.PluginTimeout,
		ExecPipe:              r.ExecPipe,
		Experimental:          r.Experimental,
	}

	return nil
}

// TODO: remove when we remove the deprecated array format for templates
//
// Deprecated: custom unmarshalling will be removed in the next version
func (c Config) MarshalYAML() (any, error) {
	aux := rawConfig{
		DataSources:           c.DataSources,
		Context:               c.Context,
		Templates:             c.Templates,
		Plugins:               c.Plugins,
		Input:                 c.Input,
		InputDir:              c.InputDir,
		InputFiles:            c.InputFiles,
		ExcludeGlob:           c.ExcludeGlob,
		ExcludeProcessingGlob: c.ExcludeProcessingGlob,
		OutputDir:             c.OutputDir,
		OutputMap:             c.OutputMap,
		OutputFiles:           c.OutputFiles,
		OutMode:               c.OutMode,
		LDelim:                c.LDelim,
		RDelim:                c.RDelim,
		MissingKey:            c.MissingKey,
		PostExec:              c.PostExec,
		PluginTimeout:         c.PluginTimeout,
		ExecPipe:              c.ExecPipe,
		Experimental:          c.Experimental,
	}

	return aux, nil
}

// mergeDataSourceMaps - use d as defaults, and override with values from o
func mergeDataSourceMaps(d, o map[string]DataSource) map[string]DataSource {
	for k, v := range o {
		c, ok := d[k]
		if ok {
			d[k] = mergeDataSources(c, v)
		} else {
			d[k] = v
		}
	}
	return d
}

// mergeDataSources - use left as default, and override with values from right
func mergeDataSources(left, right DataSource) DataSource {
	if right.URL != nil {
		left.URL = right.URL
	}
	if left.Header == nil {
		left.Header = right.Header
	} else {
		maps.Copy(left.Header, right.Header)
	}
	return left
}

// DataSource - datasource configuration
type DataSource = config.DataSource

type PluginConfig struct {
	Cmd     string
	Args    []string      `yaml:"args,omitempty"`
	Timeout time.Duration `yaml:"timeout,omitempty"`
	Pipe    bool          `yaml:"pipe,omitempty"`
}

// UnmarshalYAML - satisfy the yaml.Umarshaler interface - plugin configs can
// either be a plain string (to specify only the name), or a map with a name,
// timeout, and pipe flag.
func (p *PluginConfig) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		s := ""
		err := value.Decode(&s)
		if err != nil {
			return err
		}

		*p = PluginConfig{Cmd: s}
		return nil
	}

	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("plugin config must be a string or map")
	}

	type raw struct {
		Cmd     string
		Args    []string
		Timeout time.Duration
		Pipe    bool
	}
	r := raw{}
	err := value.Decode(&r)
	if err != nil {
		return err
	}

	*p = PluginConfig(r)

	return nil
}

// MergeFrom - use this Config as the defaults, and override it with any
// non-zero values from the other Config
//
// Note that Input/InputDir/InputFiles will override each other, as well as
// OutputDir/OutputFiles.
func (c *Config) MergeFrom(o *Config) *Config {
	switch {
	case !isZero(o.Input):
		c.Input = o.Input
		c.InputDir = ""
		c.InputFiles = nil
		c.OutputDir = ""
	case !isZero(o.InputDir):
		c.Input = ""
		c.InputDir = o.InputDir
		c.InputFiles = nil
	case !isZero(o.InputFiles):
		if len(o.InputFiles) != 1 || o.InputFiles[0] != "-" {
			c.Input = ""
			c.InputFiles = o.InputFiles
			c.InputDir = ""
			c.OutputDir = ""
		}
	}

	if !isZero(o.OutputMap) {
		c.OutputDir = ""
		c.OutputFiles = nil
		c.OutputMap = o.OutputMap
	}
	if !isZero(o.OutputDir) {
		c.OutputDir = o.OutputDir
		c.OutputFiles = nil
		c.OutputMap = ""
	}
	if !isZero(o.OutputFiles) {
		c.OutputDir = ""
		c.OutputFiles = o.OutputFiles
		c.OutputMap = ""
	}
	if !isZero(o.ExecPipe) {
		c.ExecPipe = o.ExecPipe
		c.PostExec = o.PostExec
		c.OutputFiles = o.OutputFiles
	}
	if !isZero(o.ExcludeGlob) {
		c.ExcludeGlob = o.ExcludeGlob
	}
	if !isZero(o.ExcludeProcessingGlob) {
		c.ExcludeProcessingGlob = o.ExcludeProcessingGlob
	}
	if !isZero(o.OutMode) {
		c.OutMode = o.OutMode
	}
	if !isZero(o.LDelim) {
		c.LDelim = o.LDelim
	}
	if !isZero(o.RDelim) {
		c.RDelim = o.RDelim
	}
	if c.Templates == nil {
		c.Templates = o.Templates
	} else {
		c.Templates = mergeDataSourceMaps(c.Templates, o.Templates)
	}
	if c.DataSources == nil {
		c.DataSources = o.DataSources
	} else {
		c.DataSources = mergeDataSourceMaps(c.DataSources, o.DataSources)
	}
	if c.Context == nil {
		c.Context = o.Context
	} else {
		c.Context = mergeDataSourceMaps(c.Context, o.Context)
	}
	if len(o.Plugins) > 0 {
		maps.Copy(c.Plugins, o.Plugins)
	}

	return c
}

// validate the Config
func (c Config) validate() (err error) {
	err = notTogether(
		[]string{"in", "inputFiles", "inputDir"},
		c.Input, c.InputFiles, c.InputDir)
	if err == nil {
		err = notTogether(
			[]string{"outputFiles", "outputDir", "outputMap"},
			c.OutputFiles, c.OutputDir, c.OutputMap)
	}

	if err == nil {
		err = notTogether(
			[]string{"outputDir", "outputMap", "execPipe"},
			c.OutputDir, c.OutputMap, c.ExecPipe)
	}

	if err == nil {
		err = mustTogether("outputDir", "inputDir",
			c.OutputDir, c.InputDir)
	}

	if err == nil {
		err = mustTogether("outputMap", "inputDir",
			c.OutputMap, c.InputDir)
	}

	if err == nil {
		f := len(c.InputFiles)
		if f == 0 && c.Input != "" {
			f = 1
		}
		o := len(c.OutputFiles)
		if f != o && !c.ExecPipe {
			err = fmt.Errorf("must provide same number of 'outputFiles' (%d) as 'in' or 'inputFiles' (%d) options", o, f)
		}
	}

	if err == nil {
		if c.ExecPipe && len(c.PostExec) == 0 {
			err = fmt.Errorf("execPipe may only be used with a postExec command")
		}
	}

	if err == nil {
		if c.ExecPipe && (len(c.OutputFiles) > 0 && c.OutputFiles[0] != "-") {
			err = fmt.Errorf("must not set 'outputFiles' when using 'execPipe'")
		}
	}

	if err == nil {
		missingKeyValues := []string{"", "error", "zero", "default", "invalid"}
		if !slices.Contains(missingKeyValues, c.MissingKey) {
			err = fmt.Errorf("not allowed value for the 'missing-key' flag: %s. Allowed values: %s", c.MissingKey, strings.Join(missingKeyValues, ","))
		}
	}

	return err
}

func notTogether(names []string, values ...any) error {
	found := ""
	for i, value := range values {
		if isZero(value) {
			continue
		}
		if found != "" {
			return fmt.Errorf("only one of these options is supported at a time: '%s', '%s'",
				found, names[i])
		}
		found = names[i]
	}
	return nil
}

func mustTogether(left, right string, lValue, rValue any) error {
	if !isZero(lValue) && isZero(rValue) {
		return fmt.Errorf("these options must be set together: '%s', '%s'",
			left, right)
	}

	return nil
}

func isZero(value any) bool {
	switch v := value.(type) {
	case string:
		return v == ""
	case []string:
		return len(v) == 0
	case bool:
		return !v
	default:
		return false
	}
}

// applyDefaults - any defaults changed here should be added to cmd.InitFlags as
// well for proper help/usage display.
func (c *Config) applyDefaults() {
	if c.Stdout == nil {
		c.Stdout = os.Stdout
	}
	if c.Stderr == nil {
		c.Stderr = os.Stderr
	}
	if c.Stdin == nil {
		c.Stdin = os.Stdin
	}

	if c.InputDir != "" && c.OutputDir == "" && c.OutputMap == "" {
		c.OutputDir = "."
	}
	if c.Input == "" && c.InputDir == "" && len(c.InputFiles) == 0 {
		c.InputFiles = []string{"-"}
	}
	if c.OutputDir == "" && c.OutputMap == "" && len(c.OutputFiles) == 0 {
		c.OutputFiles = []string{"-"}
	}
	if c.LDelim == "" {
		c.LDelim = "{{"
	}
	if c.RDelim == "" {
		c.RDelim = "}}"
	}
	if c.MissingKey == "" {
		c.MissingKey = "error"
	}

	if c.PluginTimeout == 0 {
		c.PluginTimeout = 5 * time.Second
	}
}

// getMode - parse an os.FileMode out of the string, and let us know if it's an override or not...
func (c *Config) getMode() (os.FileMode, bool, error) {
	modeOverride := c.OutMode != ""
	m, err := strconv.ParseUint("0"+c.OutMode, 8, 32)
	if err != nil {
		return 0, false, err
	}
	mode := iohelpers.NormalizeFileMode(os.FileMode(m))
	if mode == 0 && c.Input != "" {
		mode = iohelpers.NormalizeFileMode(0o644)
	}
	return mode, modeOverride, nil
}

// String -
func (c *Config) String() string {
	out := &strings.Builder{}
	out.WriteString("---\n")
	enc := yaml.NewEncoder(out)
	enc.SetIndent(2)

	// dereferenced copy so we can truncate input for display
	c2 := *c
	if len(c2.Input) >= 11 {
		c2.Input = c2.Input[0:8] + "..."
	}

	err := enc.Encode(c2)
	if err != nil {
		return err.Error()
	}
	return out.String()
}
