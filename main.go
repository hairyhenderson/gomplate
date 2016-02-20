package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"

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

// Getenv retrieves the value of the environment variable named by the key.
// It returns the value, which will be empty if the variable is not present.
func Getenv(key string) string {
	return os.Getenv(key)
}

// Bool converts a string to a boolean value, using strconv.ParseBool under the covers.
// Possible true values are: 1, t, T, TRUE, true, True
// All other values are considered false.
func Bool(in string) bool {
	if b, err := strconv.ParseBool(in); err == nil {
		return b
	}
	return false
}

var funcMap = template.FuncMap{
	"Getenv": Getenv,
	"getenv": Getenv,
	"Bool":   Bool,
	"bool":   Bool,
}

func createTemplate() *template.Template {
	return template.New("template").Funcs(funcMap).Option("missingkey=error")
}

// RunTemplate -
func RunTemplate(in io.Reader, out io.Writer) {
	context := &Context{}
	text, err := ioutil.ReadAll(in)
	if err != nil {
		log.Fatalf("Read failed!\n%v\n", err)
	}
	tmpl, err := createTemplate().Parse(string(text))
	if err != nil {
		log.Fatalf("Line %q: %v\n", string(text), err)
	}

	if err := tmpl.Execute(out, context); err != nil {
		panic(err)
	}
	out.Write([]byte("\n"))
}

func main() {
	RunTemplate(os.Stdin, os.Stdout)
}
