package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/hairyhenderson/gomplate/aws"

	"text/template"
)

// version
const Version = "0.1.0"

func init() {
	ver := flag.Bool("v", false, "Print version and exit")
	flag.Parse()
	if *ver {
		fmt.Println(Version)
		os.Exit(0)
	}
}

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
	out.Write([]byte("\n"))
}

// NewGomplate -
func NewGomplate() *Gomplate {
	env := &Env{}
	typeconv := &TypeConv{}
	ec2meta := &aws.Ec2Meta{}
	ec2info := aws.NewEc2Info()
	return &Gomplate{
		funcMap: template.FuncMap{
			"getenv":     env.Getenv,
			"bool":       typeconv.Bool,
			"json":       typeconv.JSON,
			"ec2meta":    ec2meta.Meta,
			"ec2dynamic": ec2meta.Dynamic,
			"ec2tag":     ec2info.Tag,
			"ec2region":  ec2meta.Region,
		},
	}
}

func main() {
	g := NewGomplate()
	g.RunTemplate(os.Stdin, os.Stdout)
}
