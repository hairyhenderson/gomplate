package gomplate

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"text/template"
	"time"

	"github.com/hairyhenderson/go-fsimpl"
	"github.com/hairyhenderson/go-fsimpl/autofs"
	"github.com/hairyhenderson/gomplate/v4/data"
	"github.com/hairyhenderson/gomplate/v4/internal/config"
	"github.com/hairyhenderson/gomplate/v4/internal/datafs"
)

// Options for template rendering.
//
// Experimental: subject to breaking changes before the next major release
type Options struct {
	// FSProvider - allows lookups of data source filesystems. Defaults to
	// [DefaultFSProvider].
	FSProvider fsimpl.FSProvider

	// Datasources - map of datasources to be read on demand when the
	// 'datasource'/'ds'/'include' functions are used.
	Datasources map[string]Datasource
	// Context - map of datasources to be read immediately and added to the
	// template's context
	Context map[string]Datasource
	// Templates - map of templates that can be referenced as nested templates
	Templates map[string]Datasource

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

	// Experimental - enable experimental features
	Experimental bool
}

// optionsFromConfig - create a set of options from the internal config struct.
// Does not set the Funcs field.
func optionsFromConfig(cfg *config.Config) Options {
	ds := make(map[string]Datasource, len(cfg.DataSources))
	for k, v := range cfg.DataSources {
		ds[k] = Datasource{
			URL:    v.URL,
			Header: v.Header,
		}
	}
	cs := make(map[string]Datasource, len(cfg.Context))
	for k, v := range cfg.Context {
		cs[k] = Datasource{
			URL:    v.URL,
			Header: v.Header,
		}
	}
	ts := make(map[string]Datasource, len(cfg.Templates))
	for k, v := range cfg.Templates {
		ts[k] = Datasource{
			URL:    v.URL,
			Header: v.Header,
		}
	}

	opts := Options{
		Datasources:  ds,
		Context:      cs,
		Templates:    ts,
		ExtraHeaders: cfg.ExtraHeaders,
		LDelim:       cfg.LDelim,
		RDelim:       cfg.RDelim,
		MissingKey:   cfg.MissingKey,
		Experimental: cfg.Experimental,
	}

	return opts
}

// Datasource - a datasource URL with optional headers
//
// Experimental: subject to breaking changes before the next major release
type Datasource struct {
	URL    *url.URL
	Header http.Header
}

// Renderer provides gomplate's core template rendering functionality.
// It should be initialized with NewRenderer.
//
// Experimental: subject to breaking changes before the next major release
type Renderer struct {
	//nolint:staticcheck
	data        *data.Data
	fsp         fsimpl.FSProvider
	nested      config.Templates
	funcs       template.FuncMap
	lDelim      string
	rDelim      string
	missingKey  string
	tctxAliases []string
}

// NewRenderer creates a new template renderer with the specified options.
// The returned renderer can be reused, but it is not (yet) safe for concurrent
// use.
//
// Experimental: subject to breaking changes before the next major release
func NewRenderer(opts Options) *Renderer {
	if Metrics == nil {
		Metrics = newMetrics()
	}

	tctxAliases := []string{}
	sources := map[string]config.DataSource{}

	for alias, ds := range opts.Context {
		tctxAliases = append(tctxAliases, alias)
		sources[alias] = config.DataSource{
			URL:    ds.URL,
			Header: ds.Header,
		}
	}
	for alias, ds := range opts.Datasources {
		sources[alias] = config.DataSource{
			URL:    ds.URL,
			Header: ds.Header,
		}
	}

	// convert the internal config.Templates to a map[string]Datasource
	// TODO: simplify when config.Templates is removed
	nested := config.Templates{}
	for alias, ds := range opts.Templates {
		nested[alias] = config.DataSource{
			URL:    ds.URL,
			Header: ds.Header,
		}
	}

	//nolint:staticcheck
	d := &data.Data{
		ExtraHeaders: opts.ExtraHeaders,
		Sources:      sources,
	}

	if opts.Funcs == nil {
		opts.Funcs = template.FuncMap{}
	}

	missingKey := opts.MissingKey
	if missingKey == "" {
		missingKey = "error"
	}

	if opts.FSProvider == nil {
		opts.FSProvider = DefaultFSProvider
	}

	return &Renderer{
		nested:      nested,
		data:        d,
		funcs:       opts.Funcs,
		tctxAliases: tctxAliases,
		lDelim:      opts.LDelim,
		rDelim:      opts.RDelim,
		missingKey:  missingKey,
		fsp:         opts.FSProvider,
	}
}

// Template contains the basic data needed to render a template with a Renderer
//
// Experimental: subject to breaking changes before the next major release
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

// RenderTemplates renders a list of templates, parsing each template's Text
// and executing it, outputting to its Writer. If a template's Writer is a
// non-os.Stdout io.Closer, it will be closed after the template is rendered.
//
// Experimental: subject to breaking changes before the next major release
func (t *Renderer) RenderTemplates(ctx context.Context, templates []Template) error {
	if datafs.FSProviderFromContext(ctx) == nil {
		ctx = datafs.ContextWithFSProvider(ctx, t.fsp)
	}

	// configure the template context with the refreshed Data value
	// only done here because the data context may have changed
	tmplctx, err := createTmplContext(ctx, t.tctxAliases, t.data)
	if err != nil {
		return err
	}

	return t.renderTemplatesWithData(ctx, templates, tmplctx)
}

func (t *Renderer) renderTemplatesWithData(ctx context.Context, templates []Template, tmplctx interface{}) error {
	// update funcs with the current context
	// only done here to ensure the context is properly set in func namespaces
	f := CreateFuncs(ctx, t.data)

	// add user-defined funcs last so they override the built-in funcs
	addToMap(f, t.funcs)

	// track some metrics for debug output
	start := time.Now()
	defer func() { Metrics.TotalRenderDuration = time.Since(start) }()
	for _, template := range templates {
		err := t.renderTemplate(ctx, template, f, tmplctx)
		if err != nil {
			return fmt.Errorf("renderTemplate: %w", err)
		}
	}
	return nil
}

func (t *Renderer) renderTemplate(ctx context.Context, template Template, f template.FuncMap, tmplctx interface{}) error {
	if template.Writer != nil {
		if wr, ok := template.Writer.(io.Closer); ok {
			defer wr.Close()
		}
	}

	tstart := time.Now()
	tmpl, err := parseTemplate(ctx, template.Name, template.Text,
		f, tmplctx, t.nested, t.lDelim, t.rDelim, t.missingKey)
	if err != nil {
		return err
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

// Render is a convenience method for rendering a single template. For more
// than one template, use RenderTemplates. If wr is a non-os.Stdout
// io.Closer, it will be closed after the template is rendered.
//
// Experimental: subject to breaking changes before the next major release
func (t *Renderer) Render(ctx context.Context, name, text string, wr io.Writer) error {
	return t.RenderTemplates(ctx, []Template{
		{Name: name, Text: text, Writer: wr},
	})
}

// DefaultFSProvider is the default filesystem provider used by gomplate
var DefaultFSProvider = sync.OnceValue[fsimpl.FSProvider](
	func() fsimpl.FSProvider {
		fsp := fsimpl.NewMux()

		// start with all go-fsimpl filesystems
		fsp.Add(autofs.FS)

		// override go-fsimpl's filefs with wdfs to handle working directories
		fsp.Add(datafs.WdFS)

		// gomplate-only filesystem
		fsp.Add(datafs.EnvFS)
		fsp.Add(datafs.StdinFS)
		fsp.Add(datafs.MergeFS)

		return fsp
	})()
