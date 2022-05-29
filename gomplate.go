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
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// gomplate -
type gomplate struct {
	tmplctx         interface{}
	funcMap         template.FuncMap
	nestedTemplates config.Templates

	leftDelim, rightDelim string
}

// runTemplate -
func (g *gomplate) runTemplate(ctx context.Context, t *tplate) error {
	tmpl, err := t.toGoTemplate(ctx, g)
	if err != nil {
		return err
	}

	wr, ok := t.target.(io.Closer)
	if ok && wr != os.Stdout {
		defer wr.Close()
	}

	return tmpl.Execute(t.target, g.tmplctx)
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

	// apply defaults before validation
	cfg.ApplyDefaults()

	err := cfg.Validate()
	if err != nil {
		return fmt.Errorf("failed to validate config: %w\n%+v", err, cfg)
	}

	d := data.FromConfig(ctx, cfg)
	log.Debug().Str("data", fmt.Sprintf("%+v", d)).Msg("created data from config")

	addCleanupHook(d.Cleanup)

	aliases := []string{}
	for k := range cfg.Context {
		aliases = append(aliases, k)
	}
	c, err := createTmplContext(ctx, aliases, d)
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
	tmpl, err := gatherTemplates(ctx, cfg, chooseNamer(cfg, g))
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

func chooseNamer(cfg *config.Config, g *gomplate) func(context.Context, string) (string, error) {
	if cfg.OutputMap == "" {
		return simpleNamer(cfg.OutputDir)
	}
	return mappingNamer(cfg.OutputMap, g)
}

func simpleNamer(outDir string) func(ctx context.Context, inPath string) (string, error) {
	return func(_ context.Context, inPath string) (string, error) {
		outPath := filepath.Join(outDir, inPath)
		return filepath.Clean(outPath), nil
	}
}

func mappingNamer(outMap string, g *gomplate) func(context.Context, string) (string, error) {
	return func(ctx context.Context, inPath string) (string, error) {
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
