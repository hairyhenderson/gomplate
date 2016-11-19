package main

import (
	"io"
	"io/ioutil"
	"log"
	"os"

	"strings"
	"text/template"

	"github.com/hairyhenderson/gomplate/aws"
	"github.com/hairyhenderson/gomplate/version"
	"github.com/urfave/cli"
)

func (g *Gomplate) createTemplate() *template.Template {
	return template.New("template").Funcs(g.funcMap).Option("missingkey=error")
}

// Gomplate -
type Gomplate struct {
	funcMap template.FuncMap
}

// RunTemplate -
func (g *Gomplate) RunTemplate(in io.Reader, out io.Writer) {
	context := &Context{}
	text, err := ioutil.ReadAll(in)
	if err != nil {
		log.Fatalf("Read failed!\n%v\n", err)
	}
	tmpl, err := g.createTemplate().Parse(string(text))
	if err != nil {
		log.Fatalf("Line %q: %v\n", string(text), err)
	}

	if err := tmpl.Execute(out, context); err != nil {
		panic(err)
	}
}

// NewGomplate -
func NewGomplate(data *Data) *Gomplate {
	env := &Env{}
	typeconv := &TypeConv{}
	ec2meta := aws.NewEc2Meta()
	ec2info := aws.NewEc2Info()
	return &Gomplate{
		funcMap: template.FuncMap{
			"getenv":     env.Getenv,
			"bool":       typeconv.Bool,
			"json":       typeconv.JSON,
			"jsonArray":  typeconv.JSONArray,
			"yaml":       typeconv.YAML,
			"yamlArray":  typeconv.YAMLArray,
			"slice":      typeconv.Slice,
			"join":       typeconv.Join,
			"ec2meta":    ec2meta.Meta,
			"ec2dynamic": ec2meta.Dynamic,
			"ec2tag":     ec2info.Tag,
			"ec2region":  ec2meta.Region,
			"title":      strings.Title,
			"toUpper":    strings.ToUpper,
			"toLower":    strings.ToLower,
			"datasource": data.Datasource,
		},
	}
}

func runTemplate(c *cli.Context) error {
	defer runCleanupHooks()
	data := NewData(c.StringSlice("datasource"))

	g := NewGomplate(data)
	g.RunTemplate(os.Stdin, os.Stdout)
	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "gomplate"
	app.Usage = "Process text files with Go templates"
	app.Version = version.Version
	app.Action = runTemplate

	app.Flags = []cli.Flag{
		cli.StringSliceFlag{
			Name:  "datasource, d",
			Usage: "Data source in alias=URL form. Specify multiple times to add multiple sources.",
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
