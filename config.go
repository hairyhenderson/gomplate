package gomplate

import (
	"os"
	"strconv"
	"strings"
)

// Config - values necessary for rendering templates with gomplate.
// Mainly for use by the CLI
type Config struct {
	Input       string
	InputFiles  []string
	InputDir    string
	ExcludeGlob []string
	OutputFiles []string
	OutputDir   string
	OutputMap   string
	OutMode     string

	DataSources       []string
	DataSourceHeaders []string
	Contexts          []string

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

// parse an os.FileMode out of the string, and let us know if it's an override or not...
func (o *Config) getMode() (os.FileMode, bool, error) {
	modeOverride := o.OutMode != ""
	m, err := strconv.ParseUint("0"+o.OutMode, 8, 32)
	if err != nil {
		return 0, false, err
	}
	mode := os.FileMode(m)
	if mode == 0 && o.Input != "" {
		mode = 0644
	}
	return mode, modeOverride, nil
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
