package gomplate

import (
	"io"
	"strings"

	"github.com/hairyhenderson/gomplate/v3/internal/config"
)

// Config - values necessary for rendering templates with gomplate.
// Mainly for use by the CLI
//
// Deprecated: this type will be phased out, internal/config.Config is used
// everywhere else, and will be exposed as API in a future version
type Config struct {
	Input       string
	InputFiles  []string
	InputDir    string
	ExcludeGlob []string
	OutputFiles []string
	OutputDir   string
	OutputMap   string
	OutMode     string
	Out         io.Writer

	DataSources       []string
	DataSourceHeaders []string
	Contexts          []string

	Plugins []string

	LDelim string
	RDelim string

	Templates []string
}

// defaults - sets any unset fields to their default value (if applicable)
func (o *Config) defaults() *Config {
	if o.OutputDir == "" {
		o.OutputDir = "."
	}
	if o.InputFiles == nil {
		o.InputFiles = []string{"-"}
	}
	if o.OutputFiles == nil {
		o.OutputFiles = []string{"-"}
	}
	if o.LDelim == "" {
		o.LDelim = "{{"
	}
	if o.RDelim == "" {
		o.RDelim = "}}"
	}
	return o
}

// nolint: gocyclo
func (o *Config) String() string {
	o.defaults()

	c := "input: "
	switch {
	case o.Input != "":
		c += "<arg>"
	case o.InputDir != "":
		c += o.InputDir
	default:
		c += strings.Join(o.InputFiles, ", ")
	}

	if len(o.ExcludeGlob) > 0 {
		c += "\nexclude: " + strings.Join(o.ExcludeGlob, ", ")
	}

	c += "\noutput: "
	switch {
	case o.InputDir != "" && o.OutputDir != ".":
		c += o.OutputDir
	case o.OutputMap != "":
		c += o.OutputMap
	default:
		c += strings.Join(o.OutputFiles, ", ")
	}

	if o.OutMode != "" {
		c += "\nchmod: " + o.OutMode
	}

	if len(o.DataSources) > 0 {
		c += "\ndatasources: " + strings.Join(o.DataSources, ", ")
	}
	if len(o.DataSourceHeaders) > 0 {
		c += "\ndatasourceheaders: " + strings.Join(o.DataSourceHeaders, ", ")
	}
	if len(o.Contexts) > 0 {
		c += "\ncontexts: " + strings.Join(o.Contexts, ", ")
	}

	if len(o.Plugins) > 0 {
		c += "\nplugins: " + strings.Join(o.Plugins, ", ")
	}

	if o.LDelim != "{{" {
		c += "\nleft_delim: " + o.LDelim
	}
	if o.RDelim != "}}" {
		c += "\nright_delim: " + o.RDelim
	}

	if len(o.Templates) > 0 {
		c += "\ntemplates: " + strings.Join(o.Templates, ", ")
	}
	return c
}

func (o *Config) toNewConfig() (*config.Config, error) {
	cfg := &config.Config{
		Input:       o.Input,
		InputFiles:  o.InputFiles,
		InputDir:    o.InputDir,
		ExcludeGlob: o.ExcludeGlob,
		OutputFiles: o.OutputFiles,
		OutputDir:   o.OutputDir,
		OutputMap:   o.OutputMap,
		OutMode:     o.OutMode,
		LDelim:      o.LDelim,
		RDelim:      o.RDelim,
		Templates:   o.Templates,
		OutWriter:   o.Out,
	}
	err := cfg.ParsePluginFlags(o.Plugins)
	if err != nil {
		return nil, err
	}
	err = cfg.ParseDataSourceFlags(o.DataSources, o.Contexts, o.DataSourceHeaders)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
