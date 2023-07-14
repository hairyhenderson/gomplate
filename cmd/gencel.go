package main

import (
	"flag"
	"fmt"

	"github.com/flanksource/gomplate/v3/gencel"
)

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		args = []string{"."}
	}
	fmt.Printf("Generating cel functions for %s\n", args)

	g := gencel.Generator{}
	g.ParsePkg(args...)
	g.Generate()
}
