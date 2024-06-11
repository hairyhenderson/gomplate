package gomplate

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"text/template"

	"github.com/hack-pad/hackpadfs"
	"github.com/hairyhenderson/gomplate/v4/internal/datafs"
	"github.com/hairyhenderson/gomplate/v4/internal/iohelpers"
	"github.com/hairyhenderson/gomplate/v4/tmpl"

	// TODO: switch back if/when fs.FS support gets merged upstream
	"github.com/hairyhenderson/xignore"
)

// ignorefile name, like .gitignore
const gomplateignore = ".gomplateignore"

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

// gatherTemplates - gather and prepare templates for rendering
func gatherTemplates(ctx context.Context, cfg *Config, outFileNamer outputNamer) ([]Template, error) {
	mode, modeOverride, err := cfg.getMode()
	if err != nil {
		return nil, err
	}

	var templates []Template

	switch {
	case cfg.Input != "":
		// open the output file - no need to close it, as it will be closed by the
		// caller later
		target, oerr := openOutFile(ctx, cfg.OutputFiles[0], 0o755, mode, modeOverride, cfg.Stdout)
		if oerr != nil {
			return nil, fmt.Errorf("openOutFile: %w", oerr)
		}

		templates = []Template{{
			// the arg-provided input string gets a special name
			Name:   "<arg>",
			Text:   cfg.Input,
			Writer: target,
		}}
	case cfg.InputDir != "":
		// input dirs presume output dirs are set too
		templates, err = walkDir(ctx, cfg, cfg.InputDir, outFileNamer, cfg.ExcludeGlob, cfg.ExcludeProcessingGlob, mode, modeOverride)
		if err != nil {
			return nil, fmt.Errorf("walkDir: %w", err)
		}
	case len(cfg.InputFiles) > 0:
		templates = make([]Template, len(cfg.InputFiles))
		for i, f := range cfg.InputFiles {
			templates[i], err = fileToTemplate(ctx, cfg, f, cfg.OutputFiles[i], mode, modeOverride)
			if err != nil {
				return nil, fmt.Errorf("fileToTemplate: %w", err)
			}
		}
	}

	return templates, nil
}

// walkDir - given an input dir `dir` and an output dir `outDir`, and a list
// of .gomplateignore and exclude globs (if any), walk the input directory and create a list of
// tplate objects, and an error, if any.
func walkDir(ctx context.Context, cfg *Config, dir string, outFileNamer outputNamer, excludeGlob []string, excludeProcessingGlob []string, mode os.FileMode, modeOverride bool) ([]Template, error) {
	dir = filepath.ToSlash(filepath.Clean(dir))

	// get a filesystem rooted in the same volume as dir (or / on non-Windows)
	fsys, err := datafs.FSysForPath(ctx, dir)
	if err != nil {
		return nil, err
	}

	// we need dir to be relative to the root of fsys
	// TODO: maybe need to do something with root here?
	_, resolvedDir, err := datafs.ResolveLocalPath(fsys, dir)
	if err != nil {
		return nil, fmt.Errorf("resolveLocalPath: %w", err)
	}

	// we need to sub the filesystem to the dir
	subfsys, err := fs.Sub(fsys, resolvedDir)
	if err != nil {
		return nil, fmt.Errorf("sub: %w", err)
	}

	// just check . because fsys is subbed to dir already
	dirStat, err := fs.Stat(subfsys, ".")
	if err != nil {
		return nil, fmt.Errorf("stat %q (%q): %w", dir, resolvedDir, err)
	}
	dirMode := dirStat.Mode()

	templates := make([]Template, 0)
	matcher := xignore.NewMatcher(subfsys)

	excludeMatches, err := matcher.Matches(".", &xignore.MatchesOptions{
		Ignorefile:    gomplateignore,
		Nested:        true, // allow nested ignorefile
		AfterPatterns: excludeGlob,
	})
	if err != nil {
		return nil, fmt.Errorf("ignore matching failed for %s: %w", dir, err)
	}

	excludeProcessingMatches, err := matcher.Matches(".", &xignore.MatchesOptions{
		// TODO: fix or replace xignore module so we can avoid attempting to read the .gomplateignore file for both exclude and excludeProcessing patterns
		Ignorefile:    gomplateignore,
		Nested:        true, // allow nested ignorefile
		AfterPatterns: excludeProcessingGlob,
	})
	if err != nil {
		return nil, fmt.Errorf("passthough matching failed for %s: %w", dir, err)
	}

	passthroughFiles := make(map[string]bool)

	for _, file := range excludeProcessingMatches.MatchedFiles {
		// files that need to be directly copied
		passthroughFiles[file] = true
	}

	// Unmatched ignorefile rules's files
	for _, file := range excludeMatches.UnmatchedFiles {
		// we want to pass an absolute (as much as possible) path to fileToTemplate
		inPath := filepath.Join(dir, file)
		inPath = filepath.ToSlash(inPath)

		// but outFileNamer expects only the filename itself
		outFile, err := outFileNamer.Name(ctx, file)
		if err != nil {
			return nil, fmt.Errorf("outFileNamer: %w", err)
		}

		_, ok := passthroughFiles[file]
		if ok {
			err = copyFileToOutDir(ctx, cfg, inPath, outFile, mode, modeOverride)
			if err != nil {
				return nil, fmt.Errorf("copyFileToOutDir: %w", err)
			}

			continue
		}

		tpl, err := fileToTemplate(ctx, cfg, inPath, outFile, mode, modeOverride)
		if err != nil {
			return nil, fmt.Errorf("fileToTemplate: %w", err)
		}

		// Ensure file parent dirs - use separate fsys for output file
		outfsys, err := datafs.FSysForPath(ctx, outFile)
		if err != nil {
			return nil, fmt.Errorf("fsysForPath: %w", err)
		}
		if err = hackpadfs.MkdirAll(outfsys, filepath.Dir(outFile), dirMode); err != nil {
			return nil, fmt.Errorf("mkdirAll %q: %w", outFile, err)
		}

		templates = append(templates, tpl)
	}

	return templates, nil
}

func readInFile(ctx context.Context, inFile string, mode os.FileMode) (source string, newmode os.FileMode, err error) {
	newmode = mode
	var b []byte

	//nolint:nestif
	if inFile == "-" {
		stdin := datafs.StdinFromContext(ctx)
		b, err = io.ReadAll(stdin)
		if err != nil {
			return source, newmode, fmt.Errorf("read from stdin: %w", err)
		}

		source = string(b)
	} else {
		var fsys fs.FS
		var si fs.FileInfo
		fsys, err = datafs.FSysForPath(ctx, inFile)
		if err != nil {
			return source, newmode, fmt.Errorf("fsysForPath: %w", err)
		}

		si, err = fs.Stat(fsys, inFile)
		if err != nil {
			return source, newmode, fmt.Errorf("stat %q: %w", inFile, err)
		}
		if mode == 0 {
			newmode = si.Mode()
		}

		// we read the file and store in memory immediately, to prevent leaking
		// file descriptors.
		b, err = fs.ReadFile(fsys, inFile)
		if err != nil {
			return source, newmode, fmt.Errorf("readAll %q: %w", inFile, err)
		}

		source = string(b)
	}
	return source, newmode, err
}

func getOutfileHandler(ctx context.Context, cfg *Config, outFile string, mode os.FileMode, modeOverride bool) (io.Writer, error) {
	// open the output file - no need to close it, as it will be closed by the
	// caller later
	target, err := openOutFile(ctx, outFile, 0o755, mode, modeOverride, cfg.Stdout)
	if err != nil {
		return nil, fmt.Errorf("openOutFile: %w", err)
	}

	return target, nil
}

func copyFileToOutDir(ctx context.Context, cfg *Config, inFile, outFile string, mode os.FileMode, modeOverride bool) error {
	sourceStr, newmode, err := readInFile(ctx, inFile, mode)
	if err != nil {
		return err
	}

	outFH, err := getOutfileHandler(ctx, cfg, outFile, newmode, modeOverride)
	if err != nil {
		return err
	}

	wr, ok := outFH.(io.Closer)
	if ok && wr != os.Stdout {
		defer wr.Close()
	}

	_, err = outFH.Write([]byte(sourceStr))
	return err
}

func fileToTemplate(ctx context.Context, cfg *Config, inFile, outFile string, mode os.FileMode, modeOverride bool) (Template, error) {
	source, newmode, err := readInFile(ctx, inFile, mode)
	if err != nil {
		return Template{}, err
	}

	target, err := getOutfileHandler(ctx, cfg, outFile, newmode, modeOverride)
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
// defer actual opening until the first non-empty write. If the file already
// exists, it will not be overwritten until the first difference is encountered.
//
// TODO: dirMode is always called with 0o755 - should either remove or make it configurable
//
//nolint:unparam
func openOutFile(ctx context.Context, filename string, dirMode, mode os.FileMode, modeOverride bool, stdout io.Writer) (out io.Writer, err error) {
	out = iohelpers.NewEmptySkipper(func() (io.Writer, error) {
		if filename == "-" {
			return iohelpers.NopCloser(stdout), nil
		}
		return createOutFile(ctx, filename, dirMode, mode, modeOverride)
	})
	return out, nil
}

func createOutFile(ctx context.Context, filename string, dirMode, mode os.FileMode, modeOverride bool) (out io.WriteCloser, err error) {
	// we only support writing out to local files for now
	fsys, err := datafs.FSysForPath(ctx, filename)
	if err != nil {
		return nil, fmt.Errorf("fsysForPath: %w", err)
	}

	mode = iohelpers.NormalizeFileMode(mode.Perm())
	if modeOverride {
		err = hackpadfs.Chmod(fsys, filename, mode)
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("failed to chmod output file %q with mode %q: %w", filename, mode, err)
		}
	}

	open := func() (out io.WriteCloser, err error) {
		// Ensure file parent dirs
		if err = hackpadfs.MkdirAll(fsys, filepath.Dir(filename), dirMode); err != nil {
			return nil, fmt.Errorf("mkdirAll %q: %w", filename, err)
		}

		f, err := hackpadfs.OpenFile(fsys, filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode)
		if err != nil {
			return out, fmt.Errorf("failed to open output file '%s' for writing: %w", filename, err)
		}
		out = f.(io.WriteCloser)

		return out, err
	}

	// if the output file already exists, we'll use a SameSkipper
	fi, err := hackpadfs.Stat(fsys, filename)
	if err != nil {
		// likely means the file just doesn't exist - further errors will be more useful
		return iohelpers.LazyWriteCloser(open), nil
	}
	if fi.IsDir() {
		// error because this is a directory
		return nil, isDirError(fi.Name())
	}

	out = iohelpers.SameSkipper(iohelpers.LazyReadCloser(func() (io.ReadCloser, error) {
		return hackpadfs.OpenFile(fsys, filename, os.O_RDONLY, mode)
	}), open)

	return out, err
}
