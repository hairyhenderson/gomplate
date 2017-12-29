package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/spf13/afero"
)

// for overriding in tests
var stdin io.ReadCloser = os.Stdin
var stdout io.WriteCloser = os.Stdout
var fs = afero.NewOsFs()

// tplate - models a tplate file...
type tplate struct {
	name     string
	target   io.Writer
	contents string
}

func (t *tplate) toGoTemplate(g *Gomplate) (*template.Template, error) {
	tmpl := template.New(t.name)
	tmpl.Option("missingkey=error")
	tmpl.Funcs(g.funcMap)
	tmpl.Delims(g.leftDelim, g.rightDelim)
	return tmpl.Parse(t.contents)
}

// loadContents - reads the template in _once_ if it hasn't yet been read. Uses the name!
func (t *tplate) loadContents() (err error) {
	if t.contents == "" {
		t.contents, err = readInput(t.name)
	}
	return err
}

func (t *tplate) addTarget(outFile string) error {
	if t.target == nil {
		target, err := openOutFile(outFile)
		if err != nil {
			return err
		}
		addCleanupHook(func() {
			// nolint: errcheck
			target.Close()
		})
		t.target = target
	}
	return nil
}

// gatherTemplates - gather and prepare input template(s) and output file(s) for rendering
func gatherTemplates(o *GomplateOpts) (templates []*tplate, err error) {
	// the arg-provided input string gets a special name
	if o.input != "" {
		templates = []*tplate{{
			name:     "<arg>",
			contents: o.input,
		}}
	}

	// input dirs presume output dirs are set too
	if o.inputDir != "" {
		o.inputFiles, o.outputFiles, err = walkDir(o.inputDir, o.outputDir, o.excludeGlob)
		if err != nil {
			return nil, err
		}
	}

	if len(templates) == 0 {
		templates = make([]*tplate, len(o.inputFiles))
		for i := range templates {
			templates[i] = &tplate{name: o.inputFiles[i]}
		}
	}

	if len(o.outputFiles) == 0 {
		o.outputFiles = []string{"-"}
	}

	for i, t := range templates {
		if err := t.loadContents(); err != nil {
			return nil, err
		}

		if err := t.addTarget(o.outputFiles[i]); err != nil {
			return nil, err
		}
	}

	return templates, nil
}

// walkDir - given an input dir `dir` and an output dir `outDir`, and a list
// of exclude globs (if any), walk the input directory and create a list of
// input and output files, and an error, if any.
func walkDir(dir, outDir string, excludeGlob []string) ([]string, []string, error) {
	dir = filepath.Clean(dir)
	outDir = filepath.Clean(outDir)

	si, err := fs.Stat(dir)
	if err != nil {
		return nil, nil, err
	}

	entries, err := afero.ReadDir(fs, dir)
	if err != nil {
		return nil, nil, err
	}

	if err = fs.MkdirAll(outDir, si.Mode()); err != nil {
		return nil, nil, err
	}

	excludes, err := executeCombinedGlob(excludeGlob)
	if err != nil {
		return nil, nil, err
	}

	inFiles := []string{}
	outFiles := []string{}
	for _, entry := range entries {
		nextInPath := filepath.Join(dir, entry.Name())
		nextOutPath := filepath.Join(outDir, entry.Name())

		if inList(excludes, nextInPath) {
			continue
		}

		if entry.IsDir() {
			i, o, err := walkDir(nextInPath, nextOutPath, excludes)
			if err != nil {
				return nil, nil, err
			}
			inFiles = append(inFiles, i...)
			outFiles = append(outFiles, o...)
		} else {
			inFiles = append(inFiles, nextInPath)
			outFiles = append(outFiles, nextOutPath)
		}
	}
	return inFiles, outFiles, nil
}

func inList(list []string, entry string) bool {
	for _, file := range list {
		if file == entry {
			return true
		}
	}

	return false
}

func openOutFile(filename string) (out io.WriteCloser, err error) {
	if filename == "-" {
		return stdout, nil
	}
	return fs.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
}

func readInput(filename string) (string, error) {
	var err error
	var inFile io.ReadCloser
	if filename == "-" {
		inFile = stdin
	} else {
		inFile, err = fs.OpenFile(filename, os.O_RDONLY, 0)
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

// takes an array of glob strings and executes it as a whole,
// returning a merged list of globbed files
func executeCombinedGlob(globArray []string) ([]string, error) {
	var combinedExcludes []string
	for _, glob := range globArray {
		excludeList, err := afero.Glob(fs, glob)
		if err != nil {
			return nil, err
		}

		combinedExcludes = append(combinedExcludes, excludeList...)
	}

	return combinedExcludes, nil
}
