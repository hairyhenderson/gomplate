// Package gomplate is a template renderer which supports a number of datasources,
// and includes hundreds of built-in functions.
package gomplate

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/hairyhenderson/gomplate/v3/data"
	"github.com/hairyhenderson/gomplate/v3/internal/config"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/spf13/afero"
)

// gomplate -
type gomplate struct {
	funcMap         template.FuncMap
	leftDelim       string
	rightDelim      string
	nestedTemplates templateAliases
	rootTemplate    *template.Template
	tmplctx         interface{}
}

// runTemplate -
func (g *gomplate) runTemplate(_ context.Context, t *tplate) error {
	tmpl, err := t.toGoTemplate(g)
	if err != nil {
		return err
	}

	// nolint: gocritic
	switch t.target.(type) {
	case io.Closer:
		if t.target != os.Stdout {
			// nolint: errcheck
			defer t.target.(io.Closer).Close()
		}
	}
	err = tmpl.Execute(t.target, g.tmplctx)
	return err
}

type templateAliases map[string]string

// newGomplate -
func newGomplate(funcMap template.FuncMap, leftDelim, rightDelim string, nested templateAliases, tctx interface{}) *gomplate {
	return &gomplate{
		leftDelim:       leftDelim,
		rightDelim:      rightDelim,
		funcMap:         funcMap,
		nestedTemplates: nested,
		tmplctx:         tctx,
	}
}

func parseTemplateArgs(templateArgs []string) (templateAliases, error) {
	nested := templateAliases{}
	for _, templateArg := range templateArgs {
		err := parseTemplateArg(templateArg, nested)
		if err != nil {
			return nil, err
		}
	}
	return nested, nil
}

func parseTemplateArg(templateArg string, ta templateAliases) error {
	parts := strings.SplitN(templateArg, "=", 2)
	pth := parts[0]
	alias := ""
	if len(parts) > 1 {
		alias = parts[0]
		pth = parts[1]
	}

	switch fi, err := fs.Stat(pth); {
	case err != nil:
		return err
	case fi.IsDir():
		files, err := afero.ReadDir(fs, pth)
		if err != nil {
			return err
		}
		prefix := pth
		if alias != "" {
			prefix = alias
		}
		for _, f := range files {
			if !f.IsDir() { // one-level only
				ta[path.Join(prefix, f.Name())] = path.Join(pth, f.Name())
			}
		}
	default:
		if alias != "" {
			ta[alias] = pth
		} else {
			ta[pth] = pth
		}
	}
	return nil
}

// RunTemplates - run all gomplate templates specified by the given configuration
//
// Deprecated: use Run instead
func RunTemplates(o *Config) error {
	cfg, err := o.toNewConfig()
	if err != nil {
		return err
	}
	return Run(context.Background(), cfg)
}

// Run all gomplate templates specified by the given configuration
func Run(ctx context.Context, cfg *config.Config) error {
	log := zerolog.Ctx(ctx)

	Metrics = newMetrics()
	defer runCleanupHooks()

	d := data.FromConfig(cfg)
	log.Debug().Str("data", fmt.Sprintf("%+v", d)).Msg("created data from config")

	addCleanupHook(d.Cleanup)
	nested, err := parseTemplateArgs(cfg.Templates)
	if err != nil {
		return err
	}
	c, err := createTmplContext(ctx, cfg.Context, d)
	if err != nil {
		return err
	}
	funcMap := Funcs(d)
	err = bindPlugins(ctx, cfg, funcMap)
	if err != nil {
		return err
	}
	g := newGomplate(funcMap, cfg.LDelim, cfg.RDelim, nested, c)

	return g.runTemplates(ctx, cfg)
}

func (g *gomplate) runTemplates(ctx context.Context, cfg *config.Config) error {
	start := time.Now()
	tmpl, err := gatherTemplates(cfg, chooseNamer(cfg, g))
	Metrics.GatherDuration = time.Since(start)
	if err != nil {
		Metrics.Errors++
		return err
	}
	Metrics.TemplatesGathered = len(tmpl)
	start = time.Now()
	defer func() { Metrics.TotalRenderDuration = time.Since(start) }()
	for _, t := range tmpl {
		tstart := time.Now()
		err := g.runTemplate(ctx, t)
		Metrics.RenderDuration[t.name] = time.Since(tstart)
		if err != nil {
			Metrics.Errors++
			return err
		}
		Metrics.TemplatesProcessed++
	}
	return nil
}

func chooseNamer(cfg *config.Config, g *gomplate) func(string) (string, error) {
	if cfg.OutputMap == "" {
		return simpleNamer(cfg.OutputDir)
	}
	return mappingNamer(cfg.OutputMap, g)
}

func simpleNamer(outDir string) func(inPath string) (string, error) {
	return func(inPath string) (string, error) {
		outPath := filepath.Join(outDir, inPath)
		return filepath.Clean(outPath), nil
	}
}

func mappingNamer(outMap string, g *gomplate) func(string) (string, error) {
	return func(inPath string) (string, error) {
		out := &bytes.Buffer{}
		t := &tplate{
			name:     "<OutputMap>",
			contents: outMap,
			target:   out,
		}
		tpl, err := t.toGoTemplate(g)
		if err != nil {
			return "", err
		}
		tctx := &tmplctx{}
		// nolint: gocritic
		switch c := g.tmplctx.(type) {
		case *tmplctx:
			for k, v := range *c {
				if k != "in" && k != "ctx" {
					(*tctx)[k] = v
				}
			}
		}
		(*tctx)["ctx"] = g.tmplctx
		(*tctx)["in"] = inPath

		err = tpl.Execute(t.target, tctx)
		if err != nil {
			return "", errors.Wrapf(err, "failed to render outputMap with ctx %+v and inPath %s", tctx, inPath)
		}

		return filepath.Clean(strings.TrimSpace(out.String())), nil
	}
}
