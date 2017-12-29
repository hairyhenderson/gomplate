package main

import (
	"io"
	"text/template"

	"github.com/hairyhenderson/gomplate/data"
)

// Gomplate -
type Gomplate struct {
	funcMap    template.FuncMap
	leftDelim  string
	rightDelim string
}

// RunTemplate -
func (g *Gomplate) RunTemplate(t *tplate) error {
	context := &Context{}
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

// NewGomplate -
func NewGomplate(d *data.Data, leftDelim, rightDelim string) *Gomplate {
	return &Gomplate{
		leftDelim:  leftDelim,
		rightDelim: rightDelim,
		funcMap:    initFuncs(d),
	}
}

func runTemplate(o *GomplateOpts) error {
	defer runCleanupHooks()
	d := data.NewData(o.dataSources, o.dataSourceHeaders)
	addCleanupHook(d.Cleanup)

	g := NewGomplate(d, o.lDelim, o.rDelim)

	tmpl, err := gatherTemplates(o)
	if err != nil {
		return err
	}
	for _, t := range tmpl {
		if err := g.RunTemplate(t); err != nil {
			return err
		}
	}
	return nil
}
