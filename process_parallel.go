package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"errors"
)

type renderResult struct {
	err     error
	inPath  string
	outPath string
}

type parallelProcessor struct{}

// == Direct input processing ========================================

func (p parallelProcessor) processInputFiles(stringTemplate string, inFiles []string, outFiles []string, g *Gomplate) error {
	inFiles, err := readInputs(stringTemplate, inFiles)
	if err != nil {
		return err
	}

	if len(outFiles) == 0 {
		outFiles = []string{"-"}
	}

	results := make(chan *renderResult)
	defer close(results)

	for n, input := range inFiles {
		go func(idx int, input string) {
			err := renderTemplate(g, input, outFiles[idx])
			results <- &renderResult{err, input, outFiles[idx]}
		}(n, input)
	}

	return p.waitAndEvaluateResults(results, len(inFiles))
}

// == Recursive input dir processing ======================================

func (p parallelProcessor) processInputDir(inputDir string, outDir string, g *Gomplate) error {
	results := make(chan *renderResult)
	defer close(results)
	nr := p.processDir(inputDir, outDir, g, results, 0)
	return p.waitAndEvaluateResults(results, nr)
}

func (p parallelProcessor) processDir(inPath string, outPath string, g *Gomplate, results chan *renderResult, nr int) int {
	inPath = filepath.Clean(inPath)
	outPath = filepath.Clean(outPath)

	// assert tha input path exists
	si, err := os.Stat(inPath)
	if err != nil {
		return p.reportError(err, inPath, outPath, results, nr)
	}

	// ensure output directory
	if err = os.MkdirAll(outPath, si.Mode()); err != nil {
		return p.reportError(err, inPath, outPath, results, nr)
	}

	// read directory
	entries, err := ioutil.ReadDir(inPath)
	if err != nil {
		return p.reportError(err, inPath, outPath, results, nr)
	}

	// process or dive in again
	for _, entry := range entries {
		nextInPath := filepath.Join(inPath, entry.Name())
		nextOutPath := filepath.Join(outPath, entry.Name())

		if entry.IsDir() {
			nr += p.processDir(nextInPath, nextOutPath, g, results, 0)
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

func (p parallelProcessor) reportError(err error, inPath string, outPath string, results chan *renderResult, nr int) int {
	go (func() {
		results <- &renderResult{err, inPath, outPath}
	})()
	return nr + 1
}

// == Thread handling ================================================

func (p parallelProcessor) waitAndEvaluateResults(results chan *renderResult, nr int) error {
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
