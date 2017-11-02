package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/zealic/xignore"
)

// Ignorefile define ignore filename
const Ignorefile = ".gomplateignore"

// == Direct input processing ========================================

func processInputFiles(stringTemplate string, input []string, output []string, g *Gomplate) error {
	input, err := readInputs(stringTemplate, input)
	if err != nil {
		return err
	}

	if len(output) == 0 {
		output = []string{"-"}
	}

	for n, input := range input {
		if err := renderTemplate(g, input, output[n]); err != nil {
			return err
		}
	}
	return nil
}

// == Recursive input dir processing ======================================
func ensureDir(dir string, mode os.FileMode) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, mode)
	}
	return nil
}

func processInputDir(input string, output string, g *Gomplate) error {
	input = filepath.Clean(input)
	ignoreMatches, err := xignore.DirMatches(input, &xignore.DirIgnoreOptions{
		IgnoreFile: Ignorefile,
	})
	if err != nil {
		return err
	}

	outputDir, err := filepath.Abs(filepath.Clean(output))
	if err != nil {
		return err
	}

	// Ensure directories
	si, err := os.Stat(input)
	if err != nil {
		return err
	}
	err = ensureDir(outputDir, si.Mode())
	if err != nil {
		return err
	}
	for _, dir := range []string(ignoreMatches.UnmatchedDirs) {
		err = ensureDir(filepath.Join(outputDir, dir), si.Mode())
		if err != nil {
			return err
		}
	}

	// Render files
	for _, tplFile := range ignoreMatches.UnmatchedFiles {
		tplFile = filepath.Join(ignoreMatches.BaseDir, tplFile)
		inString, err := readInput(tplFile)
		if err != nil {
			return err
		}

		relname, err := filepath.Rel(ignoreMatches.BaseDir, tplFile)
		if err != nil {
			return err
		}
		outputFile := filepath.Join(outputDir, relname)
		if err := renderTemplate(g, inString, outputFile); err != nil {
			return err
		}
	}

	return nil
}

// == File handling ================================================

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
		// nolint: errcheck
		defer inFile.Close()
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
