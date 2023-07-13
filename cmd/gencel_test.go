package main

import (
	"os"
	"testing"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gomplate/v3/gencel"
)

// Not really a test but just a runner so it's easier to attach a debugger.
func TestGencel(t *testing.T) {
	wd, _ := os.Getwd()
	logger.Infof("WD: %s", wd)

	args := []string{"../funcs"}
	g := gencel.Generator{}
	g.ParsePkg(args...)
	g.Generate()
}
