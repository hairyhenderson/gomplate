package gomplate

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/hairyhenderson/go-fsimpl"
	"github.com/hairyhenderson/gomplate/v4/internal/config"
	"github.com/hairyhenderson/gomplate/v4/internal/iohelpers"
	"github.com/hairyhenderson/gomplate/v4/tmpl"

	"github.com/spf13/afero"
	"github.com/zealic/xignore"
)

// ignorefile name, like .gitignore
const gomplateignore = ".gomplateignore"

// for overriding in tests
var aferoFS = afero.NewOsFs()

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

var fsProviderCtxKey = struct{}{}

// ContextWithFSProvider returns a context with the given FSProvider. Should
// only be used in tests.
func ContextWithFSProvider(ctx context.Context, fsp fsimpl.FSProvider) context.Context {
	return context.WithValue(ctx, fsProviderCtxKey, fsp)
}

// FSProviderFromContext returns the FSProvider from the context, if any
func FSProviderFromContext(ctx context.Context) fsimpl.FSProvider {
	if fsp, ok := ctx.Value(fsProviderCtxKey).(fsimpl.FSProvider); ok {
		return fsp
	}

	return nil
}

// parseTemplate - parses text as a Go template with the given name and options
func parseTemplate(ctx context.Context, name, text string, funcs template.FuncMap, tmplctx interface{}, nested config.Templates, leftDelim, rightDelim string) (tmpl *template.Template, err error) {
	tmpl = template.New(name)
	tmpl.Option("missingkey=error")

	funcMap := copyFuncMap(funcs)

	// the "tmpl" funcs get added here because they need access to the root template and context
	addTmplFuncs(funcMap, tmpl, tmplctx, name)
	tmpl.Funcs(funcMap)
	tmpl.Delims(leftDelim, rightDelim)
	_, err = tmpl.Parse(text)
	if err != nil {
		return nil, err
	}

	err = parseNestedTemplates(ctx, nested, tmpl)
	if err != nil {
		return nil, fmt.Errorf("parse nested templates: %w", err)
	}

	return tmpl, nil
}

func parseNestedTemplates(ctx context.Context, nested config.Templates, tmpl *template.Template) error {
	fsp := FSProviderFromContext(ctx)

	for alias, n := range nested {
		u := *n.URL

		fname := path.Base(u.Path)
		if strings.HasSuffix(u.Path, "/") {
			fname = "."
		}

		u.Path = path.Dir(u.Path)

		fsys, err := fsp.New(&u)
		if err != nil {
			return fmt.Errorf("filesystem provider for %q unavailable: %w", &u, err)
		}

		// inject context & header in case they're useful...
		fsys = fsimpl.WithContextFS(ctx, fsys)
		fsys = fsimpl.WithHeaderFS(n.Header, fsys)

		// valid fs.FS paths have no trailing slash
		fname = strings.TrimRight(fname, "/")

		// first determine if the template path is a directory, in which case we
		// need to load all the files in the directory (but not recursively)
		fi, err := fs.Stat(fsys, fname)
		if err != nil {
			return fmt.Errorf("stat %q: %w", fname, err)
		}

		if fi.IsDir() {
			err = parseNestedTemplateDir(ctx, fsys, alias, fname, tmpl)
		} else {
			err = parseNestedTemplate(ctx, fsys, alias, fname, tmpl)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func parseNestedTemplateDir(ctx context.Context, fsys fs.FS, alias, fname string, tmpl *template.Template) error {
	files, err := fs.ReadDir(fsys, fname)
	if err != nil {
		return fmt.Errorf("readDir %q: %w", fname, err)
	}

	for _, f := range files {
		if !f.IsDir() {
			err = parseNestedTemplate(ctx, fsys,
				path.Join(alias, f.Name()),
				path.Join(fname, f.Name()),
				tmpl,
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func parseNestedTemplate(_ context.Context, fsys fs.FS, alias, fname string, tmpl *template.Template) error {
	b, err := fs.ReadFile(fsys, fname)
	if err != nil {
		return fmt.Errorf("readFile %q: %w", fname, err)
	}

	_, err = tmpl.New(alias).Parse(string(b))
	if err != nil {
		return fmt.Errorf("parse nested template %q: %w", fname, err)
	}

	return nil
}

// gatherTemplates - gather and prepare templates for rendering
// nolint: gocyclo
func gatherTemplates(ctx context.Context, cfg *config.Config, outFileNamer func(context.Context, string) (string, error)) (templates []Template, err error) {
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

		templates = []Template{{
			Name:   "<arg>",
			Text:   cfg.Input,
			Writer: target,
		}}
	case cfg.InputDir != "":
		// input dirs presume output dirs are set too
		templates, err = walkDir(ctx, cfg, cfg.InputDir, outFileNamer, cfg.ExcludeGlob, mode, modeOverride)
		if err != nil {
			return nil, err
		}
	case cfg.Input == "":
		templates = make([]Template, len(cfg.InputFiles))
		for i := range cfg.InputFiles {
			templates[i], err = fileToTemplate(cfg, cfg.InputFiles[i], cfg.OutputFiles[i], mode, modeOverride)
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
func walkDir(ctx context.Context, cfg *config.Config, dir string, outFileNamer func(context.Context, string) (string, error), excludeGlob []string, mode os.FileMode, modeOverride bool) ([]Template, error) {
	dir = filepath.Clean(dir)

	dirStat, err := aferoFS.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("couldn't stat %s: %w", dir, err)
	}
	dirMode := dirStat.Mode()

	templates := make([]Template, 0)
	matcher := xignore.NewMatcher(aferoFS)

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

		tpl, err := fileToTemplate(cfg, inFile, outFile, mode, modeOverride)
		if err != nil {
			return nil, err
		}

		// Ensure file parent dirs
		if err = aferoFS.MkdirAll(filepath.Dir(outFile), dirMode); err != nil {
			return nil, err
		}

		templates = append(templates, tpl)
	}

	return templates, nil
}

func fileToTemplate(cfg *config.Config, inFile, outFile string, mode os.FileMode, modeOverride bool) (Template, error) {
	source := ""

	//nolint:nestif
	if inFile == "-" {
		b, err := io.ReadAll(cfg.Stdin)
		if err != nil {
			return Template{}, fmt.Errorf("failed to read from stdin: %w", err)
		}

		source = string(b)
	} else {
		si, err := aferoFS.Stat(inFile)
		if err != nil {
			return Template{}, err
		}
		if mode == 0 {
			mode = si.Mode()
		}

		// we read the file and store in memory immediately, to prevent leaking
		// file descriptors.
		f, err := aferoFS.OpenFile(inFile, os.O_RDONLY, 0)
		if err != nil {
			return Template{}, fmt.Errorf("failed to open %s: %w", inFile, err)
		}

		//nolint: errcheck
		defer f.Close()

		b, err := io.ReadAll(f)
		if err != nil {
			return Template{}, fmt.Errorf("failed to read %s: %w", inFile, err)
		}

		source = string(b)
	}

	// open the output file - no need to close it, as it will be closed by the
	// caller later
	target, err := openOutFile(outFile, 0755, mode, modeOverride, cfg.Stdout, cfg.SuppressEmpty)
	if err != nil {
		return Template{}, err
	}

	tmpl := Template{
		Name:   inFile,
		Text:   source,
		Writer: target,
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
		err = aferoFS.Chmod(filename, mode)
		if err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to chmod output file '%s' with mode %q: %w", filename, mode, err)
		}
	}

	open := func() (out io.WriteCloser, err error) {
		// Ensure file parent dirs
		if err = aferoFS.MkdirAll(filepath.Dir(filename), dirMode); err != nil {
			return nil, err
		}

		out, err = aferoFS.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode)
		if err != nil {
			return out, fmt.Errorf("failed to open output file '%s' for writing: %w", filename, err)
		}

		return out, err
	}

	// if the output file already exists, we'll use a SameSkipper
	fi, err := aferoFS.Stat(filename)
	if err != nil {
		// likely means the file just doesn't exist - further errors will be more useful
		return iohelpers.LazyWriteCloser(open), nil
	}
	if fi.IsDir() {
		// error because this is a directory
		return nil, isDirError(fi.Name())
	}

	out = iohelpers.SameSkipper(iohelpers.LazyReadCloser(func() (io.ReadCloser, error) {
		return aferoFS.OpenFile(filename, os.O_RDONLY, mode)
	}), open)

	return out, err
}
