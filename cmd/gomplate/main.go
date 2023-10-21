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
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// need to strip os.Args[0] so we only pass the actual flags
	return cmd.Main(ctx, os.Args[1:], os.Stdin, os.Stdout, os.Stderr)
}
