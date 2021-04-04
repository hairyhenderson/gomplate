// Package gomplate is a template renderer which supports a number of datasources,
// and includes hundreds of built-in functions.
package gomplate

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/hairyhenderson/gomplate/v3/data"
	"github.com/hairyhenderson/gomplate/v3/internal/config"
	"github.com/hairyhenderson/gomplate/v3/internal/datasources"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// gomplate -
type gomplate struct {
	tmplctx         interface{}
	funcMap         template.FuncMap
	nestedTemplates config.Templates
	rootTemplate    *template.Template

	leftDelim, rightDelim string
}

// runTemplate -
func (g *gomplate) runTemplate(ctx context.Context, t *tplate) error {
	tmpl, err := t.toGoTemplate(ctx, g)
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

// newGomplate -
func newGomplate(funcMap template.FuncMap, leftDelim, rightDelim string, nested config.Templates, tctx interface{}) *gomplate {
	return &gomplate{
		leftDelim:       leftDelim,
		rightDelim:      rightDelim,
		funcMap:         funcMap,
		nestedTemplates: nested,
		tmplctx:         tctx,
	}
}

// RunTemplates - run all gomplate templates specified by the given configuration
//
// Deprecated: use Run instead
func RunTemplates(o *Config) error {
	fmt.Println("Warning: RunTemplates is deprecated, use Run instead")
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

	// register datasources
	for k, v := range cfg.DataSources {
		datasources.DefaultRegistry.Register(k, v)
	}
	for k, v := range cfg.Context {
		datasources.DefaultRegistry.Register(k, v)
	}

	d := data.FromConfig(ctx, cfg)
	log.Debug().Str("data", fmt.Sprintf("%+v", d)).Msg("created data from config")

	addCleanupHook(d.Cleanup)
	c, err := createTmplContext(ctx, cfg.Context, d)
	if err != nil {
		return err
	}
	funcMap := CreateFuncs(ctx, d)
	err = bindPlugins(ctx, cfg, funcMap)
	if err != nil {
		return err
	}
	g := newGomplate(funcMap, cfg.LDelim, cfg.RDelim, cfg.Templates, c)

	return g.runTemplates(ctx, cfg)
}

func (g *gomplate) runTemplates(ctx context.Context, cfg *config.Config) error {
	start := time.Now()
	tmpl, err := gatherTemplates(cfg, chooseNamer(ctx, cfg, g))
	Metrics.GatherDuration = time.Since(start)
	if err != nil {
		Metrics.Errors++
		return fmt.Errorf("failed to gather templates for rendering: %w", err)
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
			return fmt.Errorf("failed to render template %s: %w", t.name, err)
		}
		Metrics.TemplatesProcessed++
	}
	return nil
}

func chooseNamer(ctx context.Context, cfg *config.Config, g *gomplate) func(string) (string, error) {
	if cfg.OutputMap == "" {
		return simpleNamer(cfg.OutputDir)
	}
	return mappingNamer(ctx, cfg.OutputMap, g)
}

func simpleNamer(outDir string) func(inPath string) (string, error) {
	return func(inPath string) (string, error) {
		outPath := filepath.Join(outDir, inPath)
		return filepath.Clean(outPath), nil
	}
}

func mappingNamer(ctx context.Context, outMap string, g *gomplate) func(string) (string, error) {
	return func(inPath string) (string, error) {
		out := &bytes.Buffer{}
		t := &tplate{
			name:     "<OutputMap>",
			contents: outMap,
			target:   out,
		}
		tpl, err := t.toGoTemplate(ctx, g)
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
