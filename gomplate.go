// Package gomplate is a template renderer which supports a number of datasources,
// and includes hundreds of built-in functions.
package gomplate

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/hairyhenderson/gomplate/v3/data"
	"github.com/hairyhenderson/gomplate/v3/internal/config"
)

// RunTemplates - run all gomplate templates specified by the given configuration
//
// Deprecated: use the Renderer interface instead
func RunTemplates(o *Config) error {
	cfg, err := o.toNewConfig()
	if err != nil {
		return err
	}
	return Run(context.Background(), cfg)
}

// Run all gomplate templates specified by the given configuration
func Run(ctx context.Context, cfg *config.Config) error {
	Metrics = newMetrics()
	defer runCleanupHooks()

	// apply defaults before validation
	cfg.ApplyDefaults()

	err := cfg.Validate()
	if err != nil {
		return fmt.Errorf("failed to validate config: %w\n%+v", err, cfg)
	}

	funcMap := template.FuncMap{}
	err = bindPlugins(ctx, cfg, funcMap)
	if err != nil {
		return err
	}

	// if a custom Stdin is set in the config, inject it into the context now
	ctx = data.ContextWithStdin(ctx, cfg.Stdin)

	opts := optionsFromConfig(cfg)
	opts.Funcs = funcMap
	tr := NewRenderer(opts)

	start := time.Now()

	namer := chooseNamer(cfg, tr)
	tmpl, err := gatherTemplates(ctx, cfg, namer)
	Metrics.GatherDuration = time.Since(start)
	if err != nil {
		Metrics.Errors++
		return fmt.Errorf("failed to gather templates for rendering: %w", err)
	}
	Metrics.TemplatesGathered = len(tmpl)

	err = tr.RenderTemplates(ctx, tmpl)
	if err != nil {
		return err
	}

	return nil
}

func chooseNamer(cfg *config.Config, tr *Renderer) func(context.Context, string) (string, error) {
	if cfg.OutputMap == "" {
		return simpleNamer(cfg.OutputDir)
	}
	return mappingNamer(cfg.OutputMap, tr)
}

func simpleNamer(outDir string) func(ctx context.Context, inPath string) (string, error) {
	return func(_ context.Context, inPath string) (string, error) {
		outPath := filepath.Join(outDir, inPath)
		return filepath.Clean(outPath), nil
	}
}

func mappingNamer(outMap string, tr *Renderer) func(context.Context, string) (string, error) {
	return func(ctx context.Context, inPath string) (string, error) {
		tr.data.Ctx = ctx
		tcontext, err := createTmplContext(ctx, tr.tctxAliases, tr.data)
		if err != nil {
			return "", err
		}

		// add '.in' to the template context and preserve the original context
		// in '.ctx'
		tctx := &tmplctx{}
		// nolint: gocritic
		switch c := tcontext.(type) {
		case *tmplctx:
			for k, v := range *c {
				if k != "in" && k != "ctx" {
					(*tctx)[k] = v
				}
			}
		}
		(*tctx)["ctx"] = tcontext
		(*tctx)["in"] = inPath

		out := &bytes.Buffer{}
		err = tr.renderTemplatesWithData(ctx,
			[]Template{{Name: "<OutputMap>", Text: outMap, Writer: out}}, tctx)
		if err != nil {
			return "", fmt.Errorf("failed to render outputMap with ctx %+v and inPath %s: %w", tctx, inPath, err)
		}

		return filepath.Clean(strings.TrimSpace(out.String())), nil
	}
}
