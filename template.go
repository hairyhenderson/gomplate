package gomplate

import (
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/hairyhenderson/go-fsimpl/autofs"
	"github.com/hairyhenderson/gomplate/v3/internal/config"
	"github.com/hairyhenderson/gomplate/v3/internal/iohelpers"
	"github.com/hairyhenderson/gomplate/v3/tmpl"

	"github.com/spf13/afero"
	"github.com/zealic/xignore"
)

// ignorefile name, like .gitignore
const gomplateignore = ".gomplateignore"

// for overriding in tests
var osFS = afero.NewOsFs()

// tplate - models a gomplate template file...
type tplate struct {
	name         string
	targetPath   string
	target       io.Writer
	contents     string
	mode         fs.FileMode
	modeOverride bool
}

func addTmplFuncs(f template.FuncMap, root *template.Template, ctx interface{}) {
	t := tmpl.New(root, ctx)
	tns := func() *tmpl.Template { return t }
	f["tmpl"] = tns
	f["tpl"] = t.Inline
}

func (t *tplate) toGoTemplate(fsys afero.Fs, g *gomplate) (tmpl *template.Template, err error) {
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
	for alias, nt := range g.nestedTemplates {
		fsys, err := autofs.Lookup(nt.URL.String())
		if err != nil {
			return nil, fmt.Errorf("lookup: %w", err)
		}

		f, err := fsys.Open(nt.URL.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to open file %q: %w", nt.URL.Path, err)
		}
		defer f.Close()

		b, err := io.ReadAll(f)
		if err != nil {
			return nil, fmt.Errorf("failed to read file at %s: %w", nt.URL.Path, err)
		}
		_, err = tmpl.New(alias).Parse(string(b))
		if err != nil {
			return nil, err
		}
	}
	return tmpl, nil
}

// fsysFor returns the filesystem and the relative path for an absolute path.
// Handles Windows by not assuming all paths are rooted at /
func fsysFor(path string) (fs.FS, string) {
	parts := strings.SplitAfterN(path, "/", 2)

	root := parts[0]
	if root == "" {
		root = "/"
	}

	if len(parts) > 1 {
		path = parts[1]
	}

	if path == "" {
		path = "."
	}

	return os.DirFS(root), path
}

// loadContents - reads the template
func (t *tplate) loadContents(in io.Reader) ([]byte, error) {
	if in == nil {
		f, err := osFS.Open(t.name)
		if err != nil {
			return nil, fmt.Errorf("failed to open %s: %w", t.name, err)
		}
		// nolint: errcheck
		defer f.Close()
		in = f
	}

	b, err := ioutil.ReadAll(in)
	if err != nil {
		return nil, fmt.Errorf("failed to load contents of %s: %w", t.name, err)
	}

	return b, nil
}

// gatherTemplates - gather and prepare input template(s) and output file(s) for rendering
// nolint: gocyclo
func gatherTemplates(cfg *config.Config, outFileNamer func(string) (string, error)) (templates []*tplate, err error) {
	mode, modeOverride, err := cfg.GetMode()
	if err != nil {
		return nil, err
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

// processTemplates - reads data into the given templates as necessary and opens
// outputs for writing as necessary
func processTemplates(cfg *config.Config, templates []*tplate) ([]*tplate, error) {
	for _, t := range templates {
		if t.contents == "" {
			var in io.Reader
			if t.name == "-" {
				in = cfg.Stdin
			}

			b, err := t.loadContents(in)
			if err != nil {
				return nil, err
			}

			t.contents = string(b)
		}

		if t.target == nil {
			out, err := openOutFile(cfg, t.targetPath, t.mode, t.modeOverride)
			if err != nil {
				return nil, err
			}

			t.target = out
		}
	}

	return templates, nil
}

// walkDir - given an input dir `dir` and an output dir `outDir`, and a list
// of .gomplateignore and exclude globs (if any), walk the input directory and create a list of
// tplate objects, and an error, if any.
func walkDir(dir string, outFileNamer func(string) (string, error), excludeGlob []string, mode os.FileMode, modeOverride bool) ([]*tplate, error) {
	dir = filepath.Clean(dir)

	dirStat, err := osFS.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("couldn't stat %s: %w", dir, err)
	}
	dirMode := dirStat.Mode()

	templates := make([]*tplate, 0)
	matcher := xignore.NewMatcher(osFS)

	// work around bug in xignore - a basedir of '.' doesn't work
	basedir := dir
	if basedir == "." {
		basedir, _ = os.Getwd()
	}
	matches, err := matcher.Matches(basedir, &xignore.MatchesOptions{
		Ignorefile:    gomplateignore,
		Nested:        true, // allow nested ignorefile
		AfterPatterns: excludeGlob,
	})
	if err != nil {
		return nil, fmt.Errorf("ignore matching failed for %s: %w", basedir, err)
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
			stat, perr := osFS.Stat(nextInPath)
			if perr == nil {
				fMode = stat.Mode()
			} else {
				fMode = dirMode
			}
		}

		// Ensure file parent dirs
		if err = osFS.MkdirAll(filepath.Dir(nextOutPath), dirMode); err != nil {
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
		si, err := osFS.Stat(inFile)
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

func openOutFile(cfg *config.Config, filename string, mode os.FileMode, modeOverride bool) (out io.Writer, err error) {
	if cfg.SuppressEmpty {
		out = iohelpers.NewEmptySkipper(func() (io.Writer, error) {
			if filename == "-" {
				return cfg.Stdout, nil
			}
			return createOutFile(filename, mode, modeOverride)
		})
		return out, nil
	}

	if filename == "-" {
		return cfg.Stdout, nil
	}
	return createOutFile(filename, mode, modeOverride)
}

func createOutFile(filename string, mode os.FileMode, modeOverride bool) (out io.WriteCloser, err error) {
	mode = iohelpers.NormalizeFileMode(mode.Perm())
	if modeOverride {
		err = osFS.Chmod(filename, mode)
		if err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to chmod output file '%s' with mode %q: %w", filename, mode, err)
		}
	}

	open := func() (out io.WriteCloser, err error) {
		out, err = osFS.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode)
		if err != nil {
			return out, fmt.Errorf("failed to open output file '%s' for writing: %w", filename, err)
		}

		return out, err
	}

	// if the output file already exists, we'll use a SameSkipper
	fi, err := osFS.Stat(filename)
	if err != nil {
		// likely means the file just doesn't exist - further errors will be more useful
		return iohelpers.LazyWriteCloser(open), nil
	}
	if fi.IsDir() {
		// error because this is a directory
		return nil, isDirError(fi.Name())
	}

	out = iohelpers.SameSkipper(iohelpers.LazyReadCloser(func() (io.ReadCloser, error) {
		return osFS.OpenFile(filename, os.O_RDONLY, mode)
	}), open)

	return out, err
}
