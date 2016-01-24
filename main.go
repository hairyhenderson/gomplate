package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"text/template"
)

// version
const Version = "0.0.1"

func init() {
	ver := flag.Bool("v", false, "Print version and exit")
	flag.Parse()
	if *ver {
		fmt.Println(Version)
		os.Exit(0)
	}
}

func createTemplate() *template.Template {
	return template.New("template").Option("missingkey=error")
}

// RunTemplate -
func RunTemplate(in io.Reader, out io.Writer) {
	s := bufio.NewScanner(in)
	context := &Context{}
	for s.Scan() {
		tmpl, err := createTemplate().Parse(s.Text())
		if err != nil {
			log.Fatalf("Line %q: %v\n", s.Text(), err)
		}

		if err := tmpl.Execute(out, context); err != nil {
			panic(err)
		}
		out.Write([]byte("\n"))
	}
}

func main() {
	RunTemplate(os.Stdin, os.Stdout)
}
