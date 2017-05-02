package main

import (
	"io"
	"log"
	"net/url"

	"strings"
	"text/template"

	"github.com/hairyhenderson/gomplate/aws"
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
	env := &Env{}
	typeconv := &TypeConv{}
	ec2meta := aws.NewEc2Meta()
	ec2info := aws.NewEc2Info()
	return &Gomplate{
		leftDelim:  leftDelim,
		rightDelim: rightDelim,
		funcMap: template.FuncMap{
			"getenv":           env.Getenv,
			"bool":             typeconv.Bool,
			"has":              typeconv.Has,
			"json":             typeconv.JSON,
			"jsonArray":        typeconv.JSONArray,
			"yaml":             typeconv.YAML,
			"yamlArray":        typeconv.YAMLArray,
			"slice":            typeconv.Slice,
			"join":             typeconv.Join,
			"toJSON":           typeconv.ToJSON,
			"toYAML":           typeconv.ToYAML,
			"ec2meta":          ec2meta.Meta,
			"ec2dynamic":       ec2meta.Dynamic,
			"ec2tag":           ec2info.Tag,
			"ec2region":        ec2meta.Region,
			"contains":         strings.Contains,
			"hasPrefix":        strings.HasPrefix,
			"hasSuffix":        strings.HasSuffix,
			"split":            strings.Split,
			"title":            strings.Title,
			"toUpper":          strings.ToUpper,
			"toLower":          strings.ToLower,
			"trim":             strings.Trim,
			"trimSpace":        strings.TrimSpace,
			"urlParse":         url.Parse,
			"datasource":       data.Datasource,
			"ds":               data.Datasource,
			"datasourceExists": data.DatasourceExists,
		},
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
