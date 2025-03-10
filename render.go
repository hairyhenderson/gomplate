package gomplate

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"net/http"
	"path"
	"slices"
	"strings"
	"text/template"
	"time"

	"github.com/hairyhenderson/go-fsimpl"
	"github.com/hairyhenderson/gomplate/v4/internal/datafs"
	"github.com/hairyhenderson/gomplate/v4/internal/funcs"
)

// RenderOptions - options for controlling how templates are rendered, and
// what data are available.
type RenderOptions struct {
	// Datasources - map of datasources to be read on demand when the
	// 'datasource'/'ds'/'include' functions are used.
	Datasources map[string]DataSource
	// Context - map of datasources to be read immediately and added to the
	// template's context
	Context map[string]DataSource
	// Templates - map of templates that can be referenced as nested templates
	Templates map[string]DataSource

	// Extra HTTP headers not attached to pre-defined datsources. Potentially
	// used by datasources defined in the template.
	ExtraHeaders map[string]http.Header

	// Funcs - map of functions to be added to the default template functions.
	// Duplicate functions will be overwritten by entries in this map.
	Funcs template.FuncMap

	// LeftDelim - set the left action delimiter for the template and all nested
	// templates to the specified string. Defaults to "{{"
	LDelim string
	// RightDelim - set the right action delimiter for the template and all nested
	// templates to the specified string. Defaults to "{{"
	RDelim string

	// MissingKey controls the behavior during execution if a map is indexed with a key that is not present in the map
	MissingKey string
}

// optionsFromConfig - translate the internal config struct to a RenderOptions.
// Does not set the Funcs field.
func optionsFromConfig(cfg *Config) RenderOptions {
	opts := RenderOptions{
		Datasources:  cfg.DataSources,
		Context:      cfg.Context,
		Templates:    cfg.Templates,
		ExtraHeaders: cfg.ExtraHeaders,
		LDelim:       cfg.LDelim,
		RDelim:       cfg.RDelim,
		MissingKey:   cfg.MissingKey,
	}

	return opts
}

type renderer struct {
	sr          datafs.DataSourceReader
	nested      map[string]DataSource
	funcs       template.FuncMap
	lDelim      string
	rDelim      string
	missingKey  string
	tctxAliases []string
}

// Renderer provides gomplate's core template rendering functionality.
// See [NewRenderer].
type Renderer interface {
	// RenderTemplates renders a list of templates, parsing each template's
	// Text and executing it, outputting to its Writer. If a template's Writer
	// is a non-[os.Stdout] [io.Closer], it will be closed after the template is
	// rendered.
	RenderTemplates(ctx context.Context, templates []Template) error

	// Render is a convenience method for rendering a single template. For more
	// than one template, use [Renderer.RenderTemplates]. If wr is a non-[os.Stdout]
	// [io.Closer], it will be closed after the template is rendered.
	Render(ctx context.Context, name, text string, wr io.Writer) error
}

// NewRenderer creates a new template renderer with the specified options.
// The returned renderer can be reused, but it is not (yet) safe for concurrent
// use.
//
// Experimental: subject to breaking changes before the next major release
func NewRenderer(opts RenderOptions) Renderer {
	return newRenderer(opts)
}

func newRenderer(opts RenderOptions) *renderer {
	if Metrics == nil {
		Metrics = newMetrics()
	}

	// this should be the only place where this registry is created
	reg := datafs.NewRegistry()

	tctxAliases := []string{}

	for alias, ds := range opts.Context {
		tctxAliases = append(tctxAliases, alias)
		reg.Register(alias, DataSource{
			URL:    ds.URL,
			Header: ds.Header,
		})
	}
	for alias, ds := range opts.Datasources {
		reg.Register(alias, DataSource{
			URL:    ds.URL,
			Header: ds.Header,
		})
	}

	// convert the internal Templates to a map[string]Datasource
	// TODO: simplify when Templates is removed
	nested := map[string]DataSource{}
	for alias, ds := range opts.Templates {
		nested[alias] = DataSource{
			URL:    ds.URL,
			Header: ds.Header,
		}
	}

	for k := range opts.ExtraHeaders {
		reg.AddExtraHeader(k, opts.ExtraHeaders[k])
	}

	if opts.Funcs == nil {
		opts.Funcs = template.FuncMap{}
	}

	missingKey := opts.MissingKey
	if missingKey == "" {
		missingKey = "error"
	}

	sr := datafs.NewSourceReader(reg)

	return &renderer{
		nested:      opts.Templates,
		sr:          sr,
		funcs:       opts.Funcs,
		tctxAliases: tctxAliases,
		lDelim:      opts.LDelim,
		rDelim:      opts.RDelim,
		missingKey:  missingKey,
	}
}

// Template contains the basic data needed to render a template with a Renderer
type Template struct {
	// Writer is the writer to output the rendered template to. If this writer
	// is a non-os.Stdout io.Closer, it will be closed after the template is
	// rendered.
	Writer io.Writer
	// Name is the name of the template - used for error messages
	Name string
	// Text is the template text
	Text string
}

func (r *renderer) RenderTemplates(ctx context.Context, templates []Template) error {
	if datafs.FSProviderFromContext(ctx) == nil {
		ctx = datafs.ContextWithFSProvider(ctx, DefaultFSProvider)
	}

	// configure the template context with the refreshed Data value
	// only done here because the data context may have changed
	tmplctx, err := createTmplContext(ctx, r.tctxAliases, r.sr)
	if err != nil {
		return err
	}

	return r.renderTemplatesWithData(ctx, templates, tmplctx)
}

func (r *renderer) renderTemplatesWithData(ctx context.Context, templates []Template, tmplctx any) error {
	// update funcs with the current context
	// only done here to ensure the context is properly set in func namespaces
	f := CreateFuncs(ctx)

	// add datasource funcs here because they need to share the source reader
	maps.Copy(f, funcs.CreateDataSourceFuncs(ctx, r.sr))

	// add user-defined funcs last so they override the built-in funcs
	maps.Copy(f, r.funcs)

	// track some metrics for debug output
	start := time.Now()
	defer func() { Metrics.TotalRenderDuration = time.Since(start) }()
	for _, template := range templates {
		err := r.renderTemplate(ctx, template, f, tmplctx)
		if err != nil {
			return fmt.Errorf("renderTemplate: %w", err)
		}
	}
	return nil
}

func (r *renderer) renderTemplate(ctx context.Context, template Template, f template.FuncMap, tmplctx any) error {
	if template.Writer != nil {
		if wr, ok := template.Writer.(io.Closer); ok {
			defer wr.Close()
		}
	}

	tstart := time.Now()
	tmpl, err := r.parseTemplate(ctx, template.Name, template.Text, f, tmplctx)
	if err != nil {
		return fmt.Errorf("parse template %s: %w", template.Name, err)
	}

	err = tmpl.Execute(template.Writer, tmplctx)
	Metrics.RenderDuration[template.Name] = time.Since(tstart)
	if err != nil {
		Metrics.Errors++
		return fmt.Errorf("failed to render template %s: %w", template.Name, err)
	}
	Metrics.TemplatesProcessed++

	return nil
}

func (r *renderer) Render(ctx context.Context, name, text string, wr io.Writer) error {
	return r.RenderTemplates(ctx, []Template{
		{Name: name, Text: text, Writer: wr},
	})
}

// parseTemplate - parses text as a Go template with the given name and options
func (r *renderer) parseTemplate(ctx context.Context, name, text string, funcs template.FuncMap, tmplctx any) (tmpl *template.Template, err error) {
	tmpl = template.New(name)

	missingKey := r.missingKey
	if missingKey == "" {
		missingKey = "error"
	}

	missingKeyValues := []string{"error", "zero", "default", "invalid"}
	if !slices.Contains(missingKeyValues, missingKey) {
		return nil, fmt.Errorf("not allowed value for the 'missing-key' flag: %s. Allowed values: %s", missingKey, strings.Join(missingKeyValues, ","))
	}

	tmpl.Option("missingkey=" + missingKey)

	funcMap := copyFuncMap(funcs)

	// the "tmpl" funcs get added here because they need access to the root template and context
	addTmplFuncs(funcMap, tmpl, tmplctx, name)
	tmpl.Funcs(funcMap)
	tmpl.Delims(r.lDelim, r.rDelim)
	_, err = tmpl.Parse(text)
	if err != nil {
		return nil, err
	}

	err = r.parseNestedTemplates(ctx, tmpl)
	if err != nil {
		return nil, fmt.Errorf("parse nested templates: %w", err)
	}

	return tmpl, nil
}

func (r *renderer) parseNestedTemplates(ctx context.Context, tmpl *template.Template) error {
	fsp := datafs.FSProviderFromContext(ctx)

	for alias, n := range r.nested {
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

		// TODO: maybe need to do something with root here?
		_, reldir, err := datafs.ResolveLocalPath(fsys, u.Path)
		if err != nil {
			return fmt.Errorf("resolveLocalPath: %w", err)
		}

		if reldir != "" && reldir != "." {
			fsys, err = fs.Sub(fsys, reldir)
			if err != nil {
				return fmt.Errorf("sub filesystem for %q unavailable: %w", &u, err)
			}
		}

		// inject context & header in case they're useful...
		fsys = fsimpl.WithContextFS(ctx, fsys)
		fsys = fsimpl.WithHeaderFS(n.Header, fsys)
		fsys = datafs.WithDataSourceRegistryFS(r.sr, fsys)

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

// DefaultFSProvider is the default filesystem provider used by gomplate
var DefaultFSProvider = datafs.DefaultProvider
