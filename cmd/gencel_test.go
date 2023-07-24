package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/flanksource/gomplate/v3/gencel"
)

// Not really a test but just a runner so it's easier to attach a debugger.
func testGencel(t *testing.T) {
	wd, _ := os.Getwd()
	fmt.Printf("WD: %s", wd)

	args := []string{"../funcs"}
	g := gencel.Generator{}
	g.ParsePkg(args...)
	g.Generate()
}
