package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// == Direct input processing ========================================

func processInputFiles(stringTemplate string, input []string, output []string, excludeList []string, g *Gomplate) error {
	ins, err := readInputs(stringTemplate, input)
	if err != nil {
		return err
	}

	if len(output) == 0 {
		output = []string{"-"}
	}

	for n, in := range ins {
		if err := renderTemplate(g, in, output[n]); err != nil {
			return err
		}
	}
	return nil
}

// == Recursive input dir processing ======================================

func processInputDir(input string, output string, excludeList []string, g *Gomplate) error {
	input = filepath.Clean(input)
	output = filepath.Clean(output)

	// assert tha input path exists
	si, err := os.Stat(input)
	if err != nil {
		return err
	}

	// read directory
	entries, err := ioutil.ReadDir(input)
	if err != nil {
		return err
	}

	// ensure output directory
	if err = os.MkdirAll(output, si.Mode()); err != nil {
		return err
	}

	// process or dive in again
	for _, entry := range entries {
		nextInPath := filepath.Join(input, entry.Name())
		nextOutPath := filepath.Join(output, entry.Name())

		if inList(excludeList, nextInPath) {
			continue
		}

		if entry.IsDir() {
			err := processInputDir(nextInPath, nextOutPath, excludeList, g)
			if err != nil {
				return err
			}
		} else {
			in, err := readInput(nextInPath)
			if err != nil {
				return err
			}
			if err := renderTemplate(g, in, nextOutPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func inList(list []string, entry string) bool {
	for _, file := range list {
		if file == entry {
			return true
		}
	}

	return false
}

// == File handling ================================================

func readInputs(inString string, files []string) ([]*input, error) {
	if inString != "" {
		return []*input{{
			name:     "<arg>",
			contents: inString,
		}}, nil
	}
	if len(files) == 0 {
		files = []string{"-"}
	}
	ins := make([]*input, len(files))

	for n, filename := range files {
		in, err := readInput(filename)
		if err != nil {
			return nil, err
		}
		ins[n] = in
	}
	return ins, nil
}

func readInput(filename string) (*input, error) {
	var err error
	var inFile *os.File
	if filename == "-" {
		inFile = os.Stdin
	} else {
		inFile, err = os.Open(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to open %s\n%v", filename, err)
		}
		// nolint: errcheck
		defer inFile.Close()
	}
	bytes, err := ioutil.ReadAll(inFile)
	if err != nil {
		err = fmt.Errorf("read failed for %s\n%v", filename, err)
		return nil, err
	}
	in := &input{
		name:     filename,
		contents: string(bytes),
	}
	return in, nil
}

func openOutFile(filename string) (out *os.File, err error) {
	if filename == "-" {
		return os.Stdout, nil
	}
	return os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
}
