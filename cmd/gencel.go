package main

import (
	"flag"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gomplate/v3/gencel"
)

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		args = []string{"."}
	}
	logger.Infof("Generating cel functions for %s\n", args)

	g := gencel.Generator{}
	g.ParsePkg(args...)
	g.Generate()
}
