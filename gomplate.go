package gomplate

import (
	"io"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/hairyhenderson/gomplate/data"
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
	OutMode     string

	DataSources       []string
	DataSourceHeaders []string

	PluginSources []string

	LDelim string
	RDelim string
}

// parse an os.FileMode out of the string, and let us know if it's an override or not...
func (o *Config) getMode() (os.FileMode, bool, error) {
	modeOverride := o.OutMode != ""
	m, err := strconv.ParseUint("0"+o.OutMode, 8, 32)
	if err != nil {
		return 0, false, err
	}
	mode := os.FileMode(m)
	return mode, modeOverride, nil
}

// nolint: gocyclo
func (o *Config) String() string {
	c := "input: "
	if o.Input != "" {
		c += "<arg>"
	} else if o.InputDir != "" {
		c += o.InputDir
	} else if len(o.InputFiles) > 0 {
		c += strings.Join(o.InputFiles, ", ")
	}

	if len(o.ExcludeGlob) > 0 {
		c += "\nexclude: " + strings.Join(o.ExcludeGlob, ", ")
	}

	c += "\noutput: "
	if o.InputDir != "" && o.OutputDir != "." {
		c += o.OutputDir
	} else if len(o.OutputFiles) > 0 {
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

	if o.LDelim != "{{" {
		c += "\nleft_delim: " + o.LDelim
	}
	if o.RDelim != "}}" {
		c += "\nright_delim: " + o.RDelim
	}
	return c
}

// gomplate -
type gomplate struct {
	funcMap    template.FuncMap
	leftDelim  string
	rightDelim string
}

// runTemplate -
func (g *gomplate) runTemplate(t *tplate) error {
	context := &context{}
	tmpl, err := t.toGoTemplate(g)
	if err != nil {
		return err
	}

	switch t.target.(type) {
	case io.Closer:
		if t.target != os.Stdout {
			// nolint: errcheck
			defer t.target.(io.Closer).Close()
		}
	}
	err = tmpl.Execute(t.target, context)
	return err
}

// newGomplate -
func newGomplate(d *data.Data, leftDelim, rightDelim string) *gomplate {
	return &gomplate{
		leftDelim:  leftDelim,
		rightDelim: rightDelim,
		funcMap:    Funcs(d),
	}
}

// RunTemplates - run all gomplate templates specified by the given configuration
func RunTemplates(o *Config) error {
	Metrics = newMetrics()
	defer runCleanupHooks()
	d, err := data.NewData(o.DataSources, o.DataSourceHeaders, o.PluginSources)
	if err != nil {
		return err
	}
	addCleanupHook(d.Cleanup)

	g := newGomplate(d, o.LDelim, o.RDelim)

	return g.runTemplates(o)
}

func (g *gomplate) runTemplates(o *Config) error {
	start := time.Now()
	tmpl, err := gatherTemplates(o)
	Metrics.GatherDuration = time.Since(start)
	if err != nil {
		Metrics.Errors++
		return err
	}
	Metrics.TemplatesGathered = len(tmpl)
	start = time.Now()
	defer func() { Metrics.TotalRenderDuration = time.Since(start) }()
	for _, t := range tmpl {
		tstart := time.Now()
		err := g.runTemplate(t)
		Metrics.RenderDuration[t.name] = time.Since(tstart)
		if err != nil {
			Metrics.Errors++
			return err
		}
		Metrics.TemplatesProcessed++
	}
	return nil
}
