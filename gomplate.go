package main

import (
	"io"
	"path/filepath"
	"text/template"

	"github.com/hairyhenderson/gomplate/data"
)

func (g *Gomplate) createTemplate(name string) *template.Template {
	return template.New(name).Funcs(g.funcMap).Option("missingkey=error")
}

// Gomplate -
type Gomplate struct {
	funcMap    template.FuncMap
	leftDelim  string
	rightDelim string
}

// RunTemplate -
func (g *Gomplate) RunTemplate(in *input, out io.Writer) error {
	context := &Context{}
	tmpl, err := g.createTemplate(in.name).Delims(g.leftDelim, g.rightDelim).Parse(in.contents)
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

// input - models an input file...
type input struct {
	name     string
	contents string
}

func runTemplate(o *GomplateOpts) error {
	defer runCleanupHooks()
	d := data.NewData(o.dataSources, o.dataSourceHeaders)
	addCleanupHook(d.Cleanup)

	g := NewGomplate(d, o.lDelim, o.rDelim)

	excludeList, err := executeCombinedGlob(o.excludeGlob)
	if err != nil {
		return err
	}

	if o.inputDir != "" {
		return processInputDir(o.inputDir, o.outputDir, excludeList, g)
	}

	return processInputFiles(o.input, o.inputFiles, o.outputFiles, excludeList, g)
}

// Called from process.go ...
func renderTemplate(g *Gomplate, in *input, outPath string) error {
	outFile, err := openOutFile(outPath)
	if err != nil {
		return err
	}
	// nolint: errcheck
	defer outFile.Close()
	err = g.RunTemplate(in, outFile)
	return err
}

// takes an array of glob strings and executes it as a whole,
// returning a merged list of globbed files
func executeCombinedGlob(globArray []string) ([]string, error) {
	var combinedExcludes []string
	for _, glob := range globArray {
		excludeList, err := filepath.Glob(glob)
		if err != nil {
			return nil, err
		}

		combinedExcludes = append(combinedExcludes, excludeList...)
	}

	return combinedExcludes, nil
}
