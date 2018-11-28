package gomplate

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/hairyhenderson/gomplate/conv"
	"github.com/hairyhenderson/gomplate/env"
	"github.com/pkg/errors"

	"github.com/spf13/afero"
)

// for overriding in tests
var stdin io.ReadCloser = os.Stdin
var fs = afero.NewOsFs()

// Stdout allows overriding the writer to use when templates are written to stdout ("-").
var Stdout io.WriteCloser = os.Stdout

// tplate - models a gomplate template file...
type tplate struct {
	name         string
	targetPath   string
	target       io.Writer
	contents     string
	mode         os.FileMode
	modeOverride bool
}

func (t *tplate) toGoTemplate(g *gomplate) (tmpl *template.Template, err error) {
	if g.rootTemplate != nil {
		tmpl = g.rootTemplate.New(t.name)
	} else {
		tmpl = template.New(t.name)
		g.rootTemplate = tmpl
	}
	tmpl.Option("missingkey=error")
	g.funcMap["tpl"] = g.tpl
	tmpl.Funcs(g.funcMap)
	tmpl.Delims(g.leftDelim, g.rightDelim)
	_, err = tmpl.Parse(t.contents)
	if err != nil {
		return nil, err
	}
	for alias, path := range g.nestedTemplates {
		// nolint: gosec
		b, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, err
		}
		_, err = tmpl.New(alias).Parse(string(b))
		if err != nil {
			return nil, err
		}
	}
	return tmpl, nil
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
		t.target, err = openOutFile(t.targetPath, t.mode, t.modeOverride)
	}
	return err
}

// gatherTemplates - gather and prepare input template(s) and output file(s) for rendering
// nolint: gocyclo
func gatherTemplates(o *Config) (templates []*tplate, err error) {
	mode, modeOverride, err := o.getMode()
	if err != nil {
		return nil, err
	}

	// the arg-provided input string gets a special name
	if o.Input != "" {
		if mode == 0 {
			mode = 0644
		}
		templates = []*tplate{{
			name:         "<arg>",
			contents:     o.Input,
			mode:         mode,
			modeOverride: modeOverride,
		}}

		if len(o.OutputFiles) == 1 {
			templates[0].targetPath = o.OutputFiles[0]
		}
	}

	// input dirs presume output dirs are set too
	if o.InputDir != "" {
		templates, err = walkDir(o.InputDir, o.OutputDir, o.ExcludeGlob, mode, modeOverride)
		if err != nil {
			return nil, err
		}
	} else if len(o.InputFiles) > 0 && o.Input == "" {
		templates = make([]*tplate, len(o.InputFiles))
		for i := range o.InputFiles {
			templates[i], err = fileToTemplates(o.InputFiles[i], o.OutputFiles[i], mode, modeOverride)
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
func walkDir(dir, outDir string, excludeGlob []string, mode os.FileMode, modeOverride bool) ([]*tplate, error) {
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
			t, err := walkDir(nextInPath, nextOutPath, excludes, mode, modeOverride)
			if err != nil {
				return nil, err
			}
			templates = append(templates, t...)
		} else {
			if mode == 0 {
				mode = entry.Mode()
			}
			templates = append(templates, &tplate{
				name:         nextInPath,
				targetPath:   nextOutPath,
				mode:         mode,
				modeOverride: modeOverride,
			})
		}
	}
	return templates, nil
}

func fileToTemplates(inFile, outFile string, mode os.FileMode, modeOverride bool) (*tplate, error) {
	if inFile != "-" {
		si, err := fs.Stat(inFile)
		if err != nil {
			return nil, err
		}
		if mode == 0 {
			mode = si.Mode()
		}
	}
	tmpl := &tplate{
		name:         inFile,
		targetPath:   outFile,
		mode:         mode,
		modeOverride: modeOverride,
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

func openOutFile(filename string, mode os.FileMode, modeOverride bool) (out io.WriteCloser, err error) {
	if conv.ToBool(env.Getenv("GOMPLATE_SUPPRESS_EMPTY", "false")) {
		out = newEmptySkipper(func() (io.WriteCloser, error) {
			if filename == "-" {
				return Stdout, nil
			}
			return createOutFile(filename, mode, modeOverride)
		})
		return out, nil
	}

	if filename == "-" {
		return Stdout, nil
	}
	return createOutFile(filename, mode, modeOverride)
}

func createOutFile(filename string, mode os.FileMode, modeOverride bool) (out io.WriteCloser, err error) {
	out, err = fs.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode.Perm())
	if err != nil {
		return out, err
	}
	if modeOverride {
		err = fs.Chmod(filename, mode.Perm())
	}
	return out, err
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

// emptySkipper is a io.WriteCloser wrapper that will only start writing once a
// non-whitespace byte has been encountered. The writer must be provided by the
// `open` func
type emptySkipper struct {
	open func() (io.WriteCloser, error)

	// internal
	w   io.WriteCloser
	buf *bytes.Buffer
	nw  bool
}

func newEmptySkipper(open func() (io.WriteCloser, error)) *emptySkipper {
	return &emptySkipper{
		w:    nil,
		buf:  &bytes.Buffer{},
		nw:   false,
		open: open,
	}
}

func (f *emptySkipper) Write(p []byte) (n int, err error) {
	if !f.nw {
		if allWhitespace(p) {
			// buffer the whitespace
			return f.buf.Write(p)
		}

		// first time around, so open the writer
		f.nw = true
		f.w, err = f.open()
		if err != nil {
			return 0, err
		}
		if f.w == nil {
			return 0, errors.New("nil writer returned by open")
		}
		// empty the buffer into the wrapped writer
		_, err = f.buf.WriteTo(f.w)
		if err != nil {
			return 0, err
		}
	}

	return f.w.Write(p)
}

func (f *emptySkipper) Close() error {
	if f.w != nil {
		return f.w.Close()
	}
	return nil
}

func allWhitespace(p []byte) bool {
	for _, b := range p {
		if b == ' ' || b == '\t' || b == '\n' || b == '\r' || b == '\v' {
			continue
		}
		return false
	}
	return true
}
