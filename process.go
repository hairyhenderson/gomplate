package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"io"

	"github.com/urfave/cli"
)

// == Direct input processing ========================================

func processInputFiles(c *cli.Context, g *Gomplate) error {
	inputs, err := readInputs(c.String("in"), c.StringSlice("file"))
	if err != nil {
		return err
	}

	outputs := c.StringSlice("out")
	if len(outputs) == 0 {
		outputs = []string{"-"}
	}

	for n, input := range inputs {
		if err := renderTemplate(g, input, outputs[n]); err != nil {
			return err
		}
	}
	return nil
}

// == Recursive input dir processing ======================================

func processInputDir(c *cli.Context, g *Gomplate) error {
	inputDir := c.String("input-dir")
	outDir := c.String("output-dir")
	if _, err := assertDir(outDir); err != nil {
		return err
	}
	return processDir(g, inputDir, outDir)
}

func processDir(g *Gomplate, inPath string, outPath string) error {
	inPath = filepath.Clean(inPath)
	outPath = filepath.Clean(outPath)

	// assert tha input path exists
	si, err := assertDir(inPath)
	if err != nil {
		return err
	}

	// ensure output directory
	if err = os.MkdirAll(outPath, si.Mode()); err != nil {
		return err
	}

	// read directory
	entries, err := ioutil.ReadDir(inPath)
	if err != nil {
		return err
	}

	// process or dive in again
	for _, entry := range entries {
		nextInPath := filepath.Join(inPath, entry.Name())
		nextOutPath := filepath.Join(outPath, entry.Name())

		if entry.IsDir() {
			err := processDir(g, nextInPath, nextOutPath)
			if err != nil {
				return err
			}
		} else {
			inString, err := readInput(nextInPath)
			if err != nil {
				return err
			}
			if err := renderTemplate(g, inString, nextOutPath); err != nil {
				return err
			}
		}
	}
	return nil
}

// == File handling ================================================

func assertDir(dir string) (os.FileInfo, error) {
	si, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !si.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", dir)
	}
	return si, nil
}

func readInputs(input string, files []string) ([]string, error) {
	if input != "" {
		return []string{input}, nil
	}
	if len(files) == 0 {
		files = []string{"-"}
	}
	ins := make([]string, len(files))

	for n, filename := range files {
		inString, err := readInput(filename)
		if err != nil {
			return nil, err
		}
		ins[n] = inString
	}
	return ins, nil
}

func readInput(filename string) (string, error) {
	var err error
	var inFile *os.File
	if filename == "-" {
		inFile = os.Stdin
	} else {
		inFile, err = os.Open(filename)
		if err != nil {
			return "", fmt.Errorf("failed to open %s\n%v", filename, err)
		}
		defer checkClose(inFile, &err)
	}
	bytes, err := ioutil.ReadAll(inFile)
	if err != nil {
		err = fmt.Errorf("read failed for %s\n%v", filename, err)
		return "", err
	}
	return string(bytes), nil
}

func openOutFile(filename string) (out *os.File, err error) {
	if filename == "-" {
		return os.Stdout, nil
	}
	return os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
}

func checkClose(c io.Closer, err *error) {
	cerr := c.Close()
	if *err == nil {
		*err = cerr
	}
}
