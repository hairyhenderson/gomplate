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

func RunTemplate(in io.Reader, out io.Writer) {
	s := bufio.NewScanner(in)
	for s.Scan() {
		tmpl, err := template.New("template").Option("missingkey=error").Parse(s.Text())
		if err != nil {
			log.Fatalf("Line %q: %v\n", s.Text(), err)
		}

		if err := tmpl.Execute(out, &Context{}); err != nil {
			panic(err)
		}
		out.Write([]byte("\n"))
	}
}

func main() {
	RunTemplate(os.Stdin, os.Stdout)
}
