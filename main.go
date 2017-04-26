package main

import (
	"io"
	"log"
	"os"

	"strings"
	"text/template"

	"errors"

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

func runTemplate(c *cli.Context) error {
	defer runCleanupHooks()
	data := NewData(c.StringSlice("datasource"), c.StringSlice("datasource-header"))
	lDelim := c.String("left-delim")
	rDelim := c.String("right-delim")

	g := NewGomplate(data, lDelim, rDelim)

	if err := validateInOutOptions(c); err != nil {
		return err
	}

	inputDir := c.String("input-dir")
	if inputDir != "" {
		return processInputDir(inputDir, getOutputDir(c), c.String("chown"), g)
	}

	return processInputFiles(c.String("in"), c.StringSlice("file"), c.StringSlice("out"), c.String("chown"), g)
}
func getOutputDir(c *cli.Context) string {
	out := c.String("output-dir")
	if out != "" {
		return out
	}
	return "."
}

// Called from process.go ...
func renderTemplate(g *Gomplate, inString string, outPath string, owner string) error {
	outFile, err := openOutFile(outPath)
	if err != nil {
		return err
	}
	defer checkClose(outFile, &err)
	g.RunTemplate(inString, outFile)
	if owner != "" {
		if err := chown(outPath, owner); err != nil {
			return err
		}
	}
	return nil
}

func validateInOutOptions(c *cli.Context) error {
	if c.String("input-dir") != "" {
		if c.String("in") != "" || len(c.StringSlice("file")) != 0 {
			return errors.New("--input-dir can not be used together with --in or --file")
		}
	}
	if c.String("output-dir") != "" {
		if len(c.StringSlice("out")) != 0 {
			return errors.New("--out can not be used together with --output-dir")
		}
		if c.String("input-dir") == "" {
			return errors.New("--input-dir must be set when --output-dir is set")
		}
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
			Usage: "Template file to process. Omit to use standard input (-), or use --in or --input-dir",
		},
		cli.StringFlag{
			Name:  "in, i",
			Usage: "Template string to process (alternative to --file and --input-dir)",
		},
		cli.StringFlag{
			Name:  "input-dir",
			Usage: "Directory which is examined recursively for templates (alternative to --file and --in)",
		},
		cli.StringSliceFlag{
			Name:  "out, o",
			Usage: "Output file name. Omit to use standard output (-).",
		},
		cli.StringFlag{
			Name:  "output-dir",
			Usage: "Directory to store the processed templates. Only used for --input-dir",
		},
		cli.StringFlag{
			Name:  "chown",
			Usage: "Owner (and group) of the generated files in the chown format. Only used for --out or --output-dir",
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
