package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

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
	return &Gomplate{
		funcMap: template.FuncMap{
			"Getenv": env.Getenv,
			"getenv": env.Getenv,
			"Bool":   typeconv.Bool,
			"bool":   typeconv.Bool,
		},
	}
}

func main() {
	g := NewGomplate()
	g.RunTemplate(os.Stdin, os.Stdout)
}
