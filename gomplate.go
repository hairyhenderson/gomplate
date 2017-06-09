package main

import (
	"io"
	"log"
	"text/template"
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
func (g *Gomplate) RunTemplate(text string, out io.Writer) {
	context := &Context{}
	tmpl, err := g.createTemplate().Delims(g.leftDelim, g.rightDelim).Parse(text)
	if err != nil {
		log.Fatalf("Line %q: %v\n", text, err)
	}

	if err := tmpl.Execute(out, context); err != nil {
		panic(err)
	}
}

// NewGomplate -
func NewGomplate(data *Data, leftDelim, rightDelim string) *Gomplate {
	return &Gomplate{
		leftDelim:  leftDelim,
		rightDelim: rightDelim,
		funcMap:    initFuncs(data),
	}
}

func runTemplate(o *GomplateOpts) error {
	defer runCleanupHooks()
	data := NewData(o.dataSources, o.dataSourceHeaders)

	g := NewGomplate(data, o.lDelim, o.rDelim)

	if o.inputDir != "" {
		return processInputDir(o.inputDir, o.outputDir, g)
	}

	return processInputFiles(o.input, o.inputFiles, o.outputFiles, g)
}

// Called from process.go ...
func renderTemplate(g *Gomplate, inString string, outPath string) error {
	outFile, err := openOutFile(outPath)
	if err != nil {
		return err
	}
	// nolint: errcheck
	defer outFile.Close()
	g.RunTemplate(inString, outFile)
	return nil
}
