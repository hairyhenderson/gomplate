package gomplate

import (
	"context"
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
var fs = afero.NewOsFs()

// tplate - models a gomplate template file...
type tplate struct {
	name         string
	target       io.Writer
	contents     string
	mode         os.FileMode
	modeOverride bool
}

func addTmplFuncs(f template.FuncMap, root *template.Template, tctx interface{}, path string) {
	t := tmpl.New(root, tctx, path)
	tns := func() *tmpl.Template { return t }
	f["tmpl"] = tns
	f["tpl"] = t.Inline
}

// copyFuncMap - copies the template.FuncMap into a new map so we can modify it
// without affecting the original
func copyFuncMap(funcMap template.FuncMap) template.FuncMap {
	if funcMap == nil {
		return nil
	}

	newFuncMap := make(template.FuncMap, len(funcMap))
	for k, v := range funcMap {
		newFuncMap[k] = v
	}
	return newFuncMap
}

func (t *tplate) toGoTemplate(ctx context.Context, g *gomplate) (tmpl *template.Template, err error) {
	tmpl = template.New(t.name)
	tmpl.Option("missingkey=error")

	funcMap := copyFuncMap(g.funcMap)

	// the "tmpl" funcs get added here because they need access to the root template and context
	addTmplFuncs(funcMap, tmpl, g.tmplctx, t.name)
	tmpl.Funcs(funcMap)
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

// loadContents - reads the template
func (t *tplate) loadContents(in io.Reader) ([]byte, error) {
	if in == nil {
		f, err := fs.OpenFile(t.name, os.O_RDONLY, 0)
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
func gatherTemplates(ctx context.Context, cfg *config.Config, outFileNamer func(context.Context, string) (string, error)) (templates []*tplate, err error) {
	mode, modeOverride, err := cfg.GetMode()
	if err != nil {
		return nil, err
	}

	switch {
	// the arg-provided input string gets a special name
	case cfg.Input != "":
		// open the output file - no need to close it, as it will be closed by the
		// caller later
		target, oerr := openOutFile(cfg.OutputFiles[0], 0755, mode, modeOverride, cfg.Stdout, cfg.SuppressEmpty)
		if oerr != nil {
			return nil, oerr
		}

		templates = []*tplate{{
			name:         "<arg>",
			contents:     cfg.Input,
			target:       target,
			mode:         mode,
			modeOverride: modeOverride,
		}}
	case cfg.InputDir != "":
		// input dirs presume output dirs are set too
		templates, err = walkDir(ctx, cfg, cfg.InputDir, outFileNamer, cfg.ExcludeGlob, mode, modeOverride)
		if err != nil {
			return nil, err
		}
	case cfg.Input == "":
		templates = make([]*tplate, len(cfg.InputFiles))
		for i := range cfg.InputFiles {
			templates[i], err = fileToTemplates(cfg, cfg.InputFiles[i], cfg.OutputFiles[i], mode, modeOverride)
			if err != nil {
				return nil, err
			}
		}
	}

	return templates, nil
}

// walkDir - given an input dir `dir` and an output dir `outDir`, and a list
// of .gomplateignore and exclude globs (if any), walk the input directory and create a list of
// tplate objects, and an error, if any.
func walkDir(ctx context.Context, cfg *config.Config, dir string, outFileNamer func(context.Context, string) (string, error), excludeGlob []string, mode os.FileMode, modeOverride bool) ([]*tplate, error) {
	dir = filepath.Clean(dir)

	dirStat, err := fs.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("couldn't stat %s: %w", dir, err)
	}
	dirMode := dirStat.Mode()

	templates := make([]*tplate, 0)
	matcher := xignore.NewMatcher(fs)

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
		inFile := filepath.Join(dir, file)
		outFile, err := outFileNamer(ctx, file)
		if err != nil {
			return nil, err
		}

		tpl, err := fileToTemplates(cfg, inFile, outFile, mode, modeOverride)
		if err != nil {
			return nil, err
		}

		// Ensure file parent dirs
		if err = fs.MkdirAll(filepath.Dir(outFile), dirMode); err != nil {
			return nil, err
		}

		templates = append(templates, tpl)
	}

	return templates, nil
}

func fileToTemplates(cfg *config.Config, inFile, outFile string, mode os.FileMode, modeOverride bool) (*tplate, error) {
	source := ""

	//nolint:nestif
	if inFile == "-" {
		b, err := io.ReadAll(cfg.Stdin)
		if err != nil {
			return nil, fmt.Errorf("failed to read from stdin: %w", err)
		}

		source = string(b)
	} else {
		si, err := fs.Stat(inFile)
		if err != nil {
			return nil, err
		}
		if mode == 0 {
			mode = si.Mode()
		}

		// we read the file and store in memory immediately, to prevent leaking
		// file descriptors.
		f, err := fs.OpenFile(inFile, os.O_RDONLY, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to open %s: %w", inFile, err)
		}

		//nolint: errcheck
		defer f.Close()

		b, err := io.ReadAll(f)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", inFile, err)
		}

		source = string(b)
	}

	// open the output file - no need to close it, as it will be closed by the
	// caller later
	target, err := openOutFile(outFile, 0755, mode, modeOverride, cfg.Stdout, cfg.SuppressEmpty)
	if err != nil {
		return nil, err
	}

	tmpl := &tplate{
		name:         inFile,
		contents:     source,
		target:       target,
		mode:         mode,
		modeOverride: modeOverride,
	}

	return tmpl, nil
}

// openOutFile returns a writer for the given file, creating the file if it
// doesn't exist yet, and creating the parent directories if necessary. Will
// defer actual opening until the first write (or the first non-empty write if
// 'suppressEmpty' is true). If the file already exists, it will not be
// overwritten until the first difference is encountered.
//
// TODO: the 'suppressEmpty' behaviour should be always enabled, in the next
// major release (v4.x).
func openOutFile(filename string, dirMode, mode os.FileMode, modeOverride bool, stdout io.Writer, suppressEmpty bool) (out io.Writer, err error) {
	if suppressEmpty {
		out = iohelpers.NewEmptySkipper(func() (io.Writer, error) {
			if filename == "-" {
				return stdout, nil
			}
			return createOutFile(filename, dirMode, mode, modeOverride)
		})
		return out, nil
	}

	if filename == "-" {
		return stdout, nil
	}
	return createOutFile(filename, dirMode, mode, modeOverride)
}

func createOutFile(filename string, dirMode, mode os.FileMode, modeOverride bool) (out io.WriteCloser, err error) {
	mode = iohelpers.NormalizeFileMode(mode.Perm())
	if modeOverride {
		err = fs.Chmod(filename, mode)
		if err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to chmod output file '%s' with mode %q: %w", filename, mode, err)
		}
	}

	open := func() (out io.WriteCloser, err error) {
		// Ensure file parent dirs
		if err = fs.MkdirAll(filepath.Dir(filename), dirMode); err != nil {
			return nil, err
		}

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
