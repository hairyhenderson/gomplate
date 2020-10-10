package gomplate

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/hairyhenderson/gomplate/v3/internal/config"
	"github.com/hairyhenderson/gomplate/v3/internal/iohelpers"
	"github.com/hairyhenderson/gomplate/v3/tmpl"

	"github.com/spf13/afero"
	"github.com/zealic/xignore"
)

// ignorefile name, like .gitignore
const gomplateignore = ".gomplateignore"

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

func addTmplFuncs(f template.FuncMap, root *template.Template, ctx interface{}) {
	t := tmpl.New(root, ctx)
	tns := func() *tmpl.Template { return t }
	f["tmpl"] = tns
	f["tpl"] = t.Inline
}

func (t *tplate) toGoTemplate(g *gomplate) (tmpl *template.Template, err error) {
	if g.rootTemplate != nil {
		tmpl = g.rootTemplate.New(t.name)
	} else {
		tmpl = template.New(t.name)
		g.rootTemplate = tmpl
	}
	tmpl.Option("missingkey=error")
	// the "tmpl" funcs get added here because they need access to the root template and context
	addTmplFuncs(g.funcMap, g.rootTemplate, g.tmplctx)
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

func (t *tplate) addTarget(cfg *config.Config) (err error) {
	if t.name == "<arg>" && t.targetPath == "" {
		t.targetPath = "-"
	}
	if t.target == nil {
		t.target, err = openOutFile(cfg, t.targetPath, t.mode, t.modeOverride)
	}
	return err
}

// gatherTemplates - gather and prepare input template(s) and output file(s) for rendering
// nolint: gocyclo
func gatherTemplates(cfg *config.Config, outFileNamer func(string) (string, error)) (templates []*tplate, err error) {
	mode, modeOverride, err := cfg.GetMode()
	if err != nil {
		return nil, err
	}

	// --exec-pipe redirects standard out to the out pipe
	if cfg.OutWriter != nil {
		Stdout = &iohelpers.NopCloser{Writer: cfg.OutWriter}
	}

	switch {
	// the arg-provided input string gets a special name
	case cfg.Input != "":
		templates = []*tplate{{
			name:         "<arg>",
			contents:     cfg.Input,
			mode:         mode,
			modeOverride: modeOverride,
			targetPath:   cfg.OutputFiles[0],
		}}
	case cfg.InputDir != "":
		// input dirs presume output dirs are set too
		templates, err = walkDir(cfg.InputDir, outFileNamer, cfg.ExcludeGlob, mode, modeOverride)
		if err != nil {
			return nil, err
		}
	case cfg.Input == "":
		templates = make([]*tplate, len(cfg.InputFiles))
		for i := range cfg.InputFiles {
			templates[i], err = fileToTemplates(cfg.InputFiles[i], cfg.OutputFiles[i], mode, modeOverride)
			if err != nil {
				return nil, err
			}
		}
	}

	return processTemplates(cfg, templates)
}

func processTemplates(cfg *config.Config, templates []*tplate) ([]*tplate, error) {
	for _, t := range templates {
		if err := t.loadContents(); err != nil {
			return nil, err
		}

		if err := t.addTarget(cfg); err != nil {
			return nil, err
		}
	}

	return templates, nil
}

// walkDir - given an input dir `dir` and an output dir `outDir`, and a list
// of .gomplateignore and exclude globs (if any), walk the input directory and create a list of
// tplate objects, and an error, if any.
func walkDir(dir string, outFileNamer func(string) (string, error), excludeGlob []string, mode os.FileMode, modeOverride bool) ([]*tplate, error) {
	dir = filepath.Clean(dir)

	dirStat, err := fs.Stat(dir)
	if err != nil {
		return nil, err
	}
	dirMode := dirStat.Mode()

	templates := make([]*tplate, 0)
	matcher := xignore.NewMatcher(fs)
	matches, err := matcher.Matches(dir, &xignore.MatchesOptions{
		Ignorefile:    gomplateignore,
		Nested:        true, // allow nested ignorefile
		AfterPatterns: excludeGlob,
	})
	if err != nil {
		return nil, err
	}

	// Unmatched ignorefile rules's files
	files := matches.UnmatchedFiles
	for _, file := range files {
		nextInPath := filepath.Join(dir, file)
		nextOutPath, err := outFileNamer(file)
		if err != nil {
			return nil, err
		}

		fMode := mode
		if mode == 0 {
			stat, perr := fs.Stat(nextInPath)
			if perr == nil {
				fMode = stat.Mode()
			} else {
				fMode = dirMode
			}
		}

		// Ensure file parent dirs
		if err = fs.MkdirAll(filepath.Dir(nextOutPath), dirMode); err != nil {
			return nil, err
		}

		templates = append(templates, &tplate{
			name:         nextInPath,
			targetPath:   nextOutPath,
			mode:         fMode,
			modeOverride: modeOverride,
		})
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

func openOutFile(cfg *config.Config, filename string, mode os.FileMode, modeOverride bool) (out io.WriteCloser, err error) {
	if cfg.SuppressEmpty {
		out = iohelpers.NewEmptySkipper(func() (io.WriteCloser, error) {
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
	mode = config.NormalizeFileMode(mode.Perm())
	if modeOverride {
		err = fs.Chmod(filename, mode)
		if err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to chmod output file '%s' with mode %q: %w", filename, mode, err)
		}
	}

	open := func() (out io.WriteCloser, err error) {
		out, err = fs.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode)
		if err != nil {
			return out, fmt.Errorf("failed to open output file '%s' for writing: %w", filename, err)
		}

		return out, err
	}

	// if the output file already exists, we'll use a SameSkipper
	fi, err := fs.Stat(filename)
	if err != nil {
		// likely means the file just doesn't exist - further errors will be more useful
		return iohelpers.LazyWriteCloser(open), nil
	}
	if fi.IsDir() {
		// error because this is a directory
		return nil, isDirError(fi.Name())
	}

	out = iohelpers.SameSkipper(iohelpers.LazyReadCloser(func() (io.ReadCloser, error) {
		return fs.OpenFile(filename, os.O_RDONLY, mode)
	}), open)

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
			return "", fmt.Errorf("failed to open %s: %w", filename, err)
		}
		// nolint: errcheck
		defer inFile.Close()
	}
	bytes, err := ioutil.ReadAll(inFile)
	if err != nil {
		err = fmt.Errorf("read failed for %s: %w", filename, err)
		return "", err
	}
	return string(bytes), nil
}
