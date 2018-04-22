package gomplate

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
var fs = afero.NewOsFs()

// Stdout allows overriding the writer to use when templates are written to stdout ("-").
var Stdout io.WriteCloser = os.Stdout

// tplate - models a gomplate template file...
type tplate struct {
	name       string
	targetPath string
	target     io.Writer
	contents   string
	mode       os.FileMode
}

func (t *tplate) toGoTemplate(g *gomplate) (*template.Template, error) {
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

func (t *tplate) addTarget() (err error) {
	if t.name == "<arg>" && t.targetPath == "" {
		t.targetPath = "-"
	}
	if t.target == nil {
		t.target, err = openOutFile(t.targetPath, t.mode)
	}
	return err
}

// gatherTemplates - gather and prepare input template(s) and output file(s) for rendering
func gatherTemplates(o *Config) (templates []*tplate, err error) {
	// the arg-provided input string gets a special name
	if o.Input != "" {
		templates = []*tplate{{
			name:     "<arg>",
			contents: o.Input,
		}}

		if len(o.OutputFiles) == 1 {
			templates[0].targetPath = o.OutputFiles[0]
		}
	}

	// input dirs presume output dirs are set too
	if o.InputDir != "" {
		templates, err = walkDir(o.InputDir, o.OutputDir, o.ExcludeGlob)
		if err != nil {
			return nil, err
		}
	} else if len(o.InputFiles) > 0 && o.Input == "" {
		templates = make([]*tplate, len(o.InputFiles))
		for i := range o.InputFiles {
			templates[i], err = fileToTemplates(o.InputFiles[i], o.OutputFiles[i])
			if err != nil {
				return nil, err
			}
		}
	}

	return processTemplates(templates)
}

func processTemplates(templates []*tplate) ([]*tplate, error) {
	for _, t := range templates {
		if err := t.loadContents(); err != nil {
			return nil, err
		}

		if err := t.addTarget(); err != nil {
			return nil, err
		}
	}

	return templates, nil
}

// walkDir - given an input dir `dir` and an output dir `outDir`, and a list
// of exclude globs (if any), walk the input directory and create a list of
// tplate objects, and an error, if any.
func walkDir(dir, outDir string, excludeGlob []string) ([]*tplate, error) {
	dir = filepath.Clean(dir)
	outDir = filepath.Clean(outDir)
	si, err := fs.Stat(dir)
	if err != nil {
		return nil, err
	}

	entries, err := afero.ReadDir(fs, dir)
	if err != nil {
		return nil, err
	}

	if err = fs.MkdirAll(outDir, si.Mode()); err != nil {
		return nil, err
	}

	excludes, err := executeCombinedGlob(excludeGlob)
	if err != nil {
		return nil, err
	}

	templates := make([]*tplate, 0)
	for _, entry := range entries {
		nextInPath := filepath.Join(dir, entry.Name())
		nextOutPath := filepath.Join(outDir, entry.Name())

		if inList(excludes, nextInPath) {
			continue
		}

		if entry.IsDir() {
			t, err := walkDir(nextInPath, nextOutPath, excludes)
			if err != nil {
				return nil, err
			}
			templates = append(templates, t...)
		} else {
			templates = append(templates, &tplate{
				name:       nextInPath,
				targetPath: nextOutPath,
				mode:       entry.Mode(),
			})
		}
	}
	return templates, nil
}

func fileToTemplates(inFile, outFile string) (*tplate, error) {
	mode := os.FileMode(0644)
	if inFile != "-" {
		si, err := fs.Stat(inFile)
		if err != nil {
			return nil, err
		}
		mode = si.Mode()
	}
	tmpl := &tplate{
		name:       inFile,
		targetPath: outFile,
		mode:       mode,
	}

	return tmpl, nil
}

func inList(list []string, entry string) bool {
	for _, file := range list {
		if file == entry {
			return true
		}
	}

	return false
}

func openOutFile(filename string, mode os.FileMode) (out io.WriteCloser, err error) {
	if filename == "-" {
		return Stdout, nil
	}
	return fs.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode.Perm())
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
