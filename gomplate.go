package main

import (
	"io"
	"path/filepath"
	"text/template"

	"github.com/hairyhenderson/gomplate/data"
)

func (g *Gomplate) createTemplate() *template.Template {
	return template.New("template").Funcs(g.funcMap).Option("missingkey=error")
}

// Gomplate -
type Gomplate struct {
	funcMap    template.FuncMap
	leftDelim  string
	rightDelim string
}

// RunTemplate -
func (g *Gomplate) RunTemplate(text string, out io.Writer) error {
	context := &Context{}
	tmpl, err := g.createTemplate().Delims(g.leftDelim, g.rightDelim).Parse(text)
	if err != nil {
		return err
	}
	err = tmpl.Execute(out, context)
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

	excludeList, err := filepath.Glob(o.excludeGlob)
	if err != nil {
		return err
	}

	if o.inputDir != "" {
		return processInputDir(o.inputDir, o.outputDir, excludeList, g)
	}

	return processInputFiles(o.input, o.inputFiles, o.outputFiles, excludeList, g)
}

// Called from process.go ...
func renderTemplate(g *Gomplate, inString string, outPath string) error {
	outFile, err := openOutFile(outPath)
	if err != nil {
		return err
	}
	// nolint: errcheck
	defer outFile.Close()
	err = g.RunTemplate(inString, outFile)
	return err
}
