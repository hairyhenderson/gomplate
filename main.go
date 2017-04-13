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
			"datasource":       data.Datasource,
			"datasourceExists": data.DatasourceExists,
		},
	}
}

func readInputs(input string, files []string) []string {
	if input != "" {
		return []string{input}
	}
	if len(files) == 0 {
		files = []string{"-"}
	}
	ins := make([]string, len(files))

	for n, filename := range files {
		var err error
		var inFile *os.File
		if filename == "-" {
			inFile = os.Stdin
		} else {
			inFile, err = os.Open(filename)
			if err != nil {
				log.Fatalf("Failed to open %s\n%v", filename, err)
			}
			defer inFile.Close() // nolint: errcheck
		}
		bytes, err := ioutil.ReadAll(inFile)
		if err != nil {
			log.Fatalf("Read failed for %s!\n%v\n", filename, err)
		}
		ins[n] = string(bytes)
	}
	return ins
}

func openOutFile(filename string) (out *os.File, err error) {
	if filename == "-" {
		return os.Stdout, nil
	}
	return os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
}

func runTemplate(c *cli.Context) error {
	defer runCleanupHooks()
	data := NewData(c.StringSlice("datasource"), c.StringSlice("datasource-header"))
	lDelim := c.String("left-delim")
	rDelim := c.String("right-delim")

	g := NewGomplate(data, lDelim, rDelim)

	inputs := readInputs(c.String("in"), c.StringSlice("file"))

	outputs := c.StringSlice("out")
	if len(outputs) == 0 {
		outputs = []string{"-"}
	}

	for n, input := range inputs {
		out, err := openOutFile(outputs[n])
		if err != nil {
			return err
		}
		defer out.Close() // nolint: errcheck
		g.RunTemplate(input, out)
	}

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
			Name:  "file, f",
			Usage: "Template file to process. Omit to use standard input (-), or use --in",
		},
		cli.StringFlag{
			Name:  "in, i",
			Usage: "Template string to process (alternative to --file)",
		},
		cli.StringSliceFlag{
			Name:  "out, o",
			Usage: "Output file name. Omit to use standard output (-).",
		},
		cli.StringSliceFlag{
			Name:  "datasource, d",
			Usage: "Data source in alias=URL form. Specify multiple times to add multiple sources.",
		},
		cli.StringSliceFlag{
			Name:  "datasource-header, H",
			Usage: "HTTP Header field in 'alias=Name: value' form to be provided on HTTP-based data sources. Multiples can be set.",
		},
		cli.StringFlag{
			Name:   "left-delim",
			Usage:  "Override the default left-delimiter `{{`",
			Value:  "{{",
			EnvVar: "GOMPLATE_LEFT_DELIM",
		},
		cli.StringFlag{
			Name:   "right-delim",
			Usage:  "Override the default right-delimiter `}}`",
			Value:  "}}",
			EnvVar: "GOMPLATE_RIGHT_DELIM",
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
