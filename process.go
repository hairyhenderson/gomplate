package main

import (
	"path/filepath"
	"io/ioutil"
	"os"
	"fmt"
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
		if err:= renderTemplate(g, input, outputs[n]); err != nil {
			return err
		}
	}
	return nil
}

// == Recursive input dir processing ======================================

func processInputDir(c *cli.Context, g *Gomplate) error {
	inputDir := c.String("input-dir")
	outDir := c.String("output-dir")
	if err := assertDirectory(outDir); err != nil {
		return err
	}
	return processDir(g, inputDir, outDir)
}

func processDir(g *Gomplate, inPath string, outPath string) error {
	inPath = filepath.Clean(inPath)
	outPath = filepath.Clean(outPath)

	// ensure input path
	si, err := os.Stat(inPath)
	if err := assertDirectory(inPath); err != nil {
		return err
	}

	// ensure output directory
	_, err = os.Stat(outPath)
	if err != nil {
	  if os.IsNotExist(err) {
		  err = os.MkdirAll(outPath, si.Mode())
	  }
		if err != nil {
			return err
		}
	}

	// read directory
	paths, err := ioutil.ReadDir(inPath)
	if err != nil {
		return err
	}

	// process or dive in again
	for _, path := range paths {
		inPath := filepath.Join(inPath, path.Name())
		outPath := filepath.Join(outPath, path.Name())

		if path.IsDir() {
			err := processDir(g, inPath, outPath)
			if err != nil {
				return err
			}
		} else {
			inString, err := readInput(inPath)
			if err != nil {
				return err
			}
			if err := renderTemplate(g, inString, outPath); err != nil {
				return err
			}
		}
	}
	return nil
}

// == Rendering ====================================================

func renderTemplate(g *Gomplate, inString string, outPath string) error {
	outFile, err := openOutFile(outPath)
	if err != nil {
		return err
	}
	defer outFile.Close()
	return g.RunTemplate(inString, outFile)
}

// == File handling ================================================

func assertDirectory(dir string) error {
	si, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !si.IsDir() {
		return fmt.Errorf("%s is not a directory", dir)
	}
	return nil
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
			return "", fmt.Errorf("Failed to open %s\n%v", filename, err)
		}
		defer inFile.Close()
	}
	bytes, err := ioutil.ReadAll(inFile)
	if err != nil {
		return "", fmt.Errorf("Read failed for %s!\n%v\n", filename, err)
	}
	return string(bytes), nil
}

func openOutFile(filename string) (out *os.File, err error) {
	if filename == "-" {
		return os.Stdout, nil
	}
	return os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
}
