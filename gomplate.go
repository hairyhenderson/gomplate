package gomplate

import (
	"io"
	"os"
	"text/template"

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

	DataSources       []string
	DataSourceHeaders []string

	LDelim string
	RDelim string
}

// gomplate -
type gomplate struct {
	funcMap    template.FuncMap
	leftDelim  string
	rightDelim string
}

type outFileInfo struct {
	name string
	mode os.FileMode
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
		// nolint: errcheck
		defer t.target.(io.Closer).Close()
	}
	err = tmpl.Execute(t.target, context)
	return err
}

// newGomplate -
func newGomplate(d *data.Data, leftDelim, rightDelim string) *gomplate {
	return &gomplate{
		leftDelim:  leftDelim,
		rightDelim: rightDelim,
		funcMap:    initFuncs(d),
	}
}

// RunTemplates - run all gomplate templates specified by the given configuration
func RunTemplates(o *Config) error {
	defer runCleanupHooks()
	d := data.NewData(o.DataSources, o.DataSourceHeaders)
	addCleanupHook(d.Cleanup)

	g := newGomplate(d, o.LDelim, o.RDelim)

	return g.runTemplates(o)
}

func (g *gomplate) runTemplates(o *Config) error {
	tmpl, err := gatherTemplates(o)
	if err != nil {
		return err
	}
	for _, t := range tmpl {
		if err := g.runTemplate(t); err != nil {
			return err
		}
	}
	return nil
}
