package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"io"

	"errors"

	"github.com/urfave/cli"
)

// == Direct input processing ========================================

type renderResult struct {
	err     error
	inPath  string
	outPath string
}

func processInputFiles(c *cli.Context, g *Gomplate) error {
	inputs, err := readInputs(c.String("in"), c.StringSlice("file"))
	if err != nil {
		return err
	}

	outputs := c.StringSlice("out")
	if len(outputs) == 0 {
		outputs = []string{"-"}
	}

	results := make(chan *renderResult)
	defer close(results)

	for n, input := range inputs {
		go func(idx int, input string) {
			if err := renderTemplate(g, input, outputs[idx]); err != nil {
				results <- &renderResult{err, input, outputs[idx]}
			}
		}(n, input)
	}

	return waitAndEvaluateResults(results, len(inputs))
}

// == Recursive input dir processing ======================================

func processInputDir(c *cli.Context, g *Gomplate) error {
	inputDir := c.String("input-dir")
	outDir := c.String("output-dir")
	if _, err := assertDir(outDir); err != nil {
		return err
	}
	results := make(chan *renderResult)
	defer close(results)
	nr := processDir(g, inputDir, outDir, results, 0)
	return waitAndEvaluateResults(results, nr)
}

func processDir(g *Gomplate, inPath string, outPath string, results chan *renderResult, nr int) int {
	inPath = filepath.Clean(inPath)
	outPath = filepath.Clean(outPath)

	// assert tha input path exists
	si, err := assertDir(inPath)
	if err != nil {
		return reportError(err, inPath, outPath, results, nr)
	}

	// ensure output directory
	if err = os.MkdirAll(outPath, si.Mode()); err != nil {
		return reportError(err, inPath, outPath, results, nr)
	}

	// read directory
	entries, err := ioutil.ReadDir(inPath)
	if err != nil {
		return reportError(err, inPath, outPath, results, nr)
	}

	// process or dive in again
	for _, entry := range entries {
		nextInPath := filepath.Join(inPath, entry.Name())
		nextOutPath := filepath.Join(outPath, entry.Name())

		if entry.IsDir() {
			nr += processDir(g, nextInPath, nextOutPath, results, 0)
		} else {
			go (func(nextInPath string, nextOutPath string) {
				inString, err := readInput(nextInPath)
				if err == nil {
					err = renderTemplate(g, inString, nextOutPath)
				}
				results <- &renderResult{err, nextInPath, nextOutPath}
			})(nextInPath, nextOutPath)
			nr++
		}
	}
	return nr
}

func reportError(err error, inPath string, outPath string, results chan *renderResult, nr int) int {
	go (func() {
		results <- &renderResult{err, inPath, outPath}
	})()
	return nr + 1
}

// == Thread handling ================================================

func waitAndEvaluateResults(results chan *renderResult, nr int) error {
	errorMsg := ""
	for i := 0; i < nr; i++ {
		result := <-results
		if result.err != nil {
			errorMsg += fmt.Sprintf("   %s --> %s : %v", result.inPath, result.outPath, result.err)
		}
	}
	if errorMsg != "" {
		return errors.New("rendering of the following templates failed:\n" + errorMsg)
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
