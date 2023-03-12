/*
The gomplate command
*/
package main

import (
	"context"
	"os"

	"github.com/hairyhenderson/gomplate/v4/internal/cmd"
)

func main() {
	exitCode := 0
	// defer the exit first, so it executes last, to let the deferred cancel run
	defer func() { os.Exit(exitCode) }()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// need to strip os.Args[0] so we only pass the actual flags
	err := cmd.Main(ctx, os.Args[1:], os.Stdin, os.Stdout, os.Stderr)
	if err != nil {
		exitCode = 1
	}
}
