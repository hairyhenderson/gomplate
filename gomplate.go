// Package gomplate is a template renderer which supports a number of datasources,
// and includes hundreds of built-in functions.
package gomplate

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/hairyhenderson/gomplate/v4/internal/datafs"
)

// Run all gomplate templates specified by the given configuration
func Run(ctx context.Context, cfg *Config) error {
	Metrics = newMetrics()

	// apply defaults before validation
	cfg.applyDefaults()

	err := cfg.validate()
	if err != nil {
		return fmt.Errorf("failed to validate config: %w\n%+v", err, cfg)
	}

	if cfg.Experimental {
		slog.SetDefault(slog.With("experimental", true))
		slog.InfoContext(ctx, "experimental functions and features enabled!")

		ctx = SetExperimental(ctx)
	}

	// bind plugins from the configuration to the funcMap
	funcMap := template.FuncMap{}
	err = bindPlugins(ctx, cfg, funcMap)
	if err != nil {
		return err
	}

	// if a custom Stdin is set in the config, inject it into the context now
	ctx = datafs.ContextWithStdin(ctx, cfg.Stdin)

	// if a custom FSProvider is set in the context, use it, otherwise inject
	// the default now - one is needed for the calls below to gatherTemplates
	// as well as the rendering itself
	if datafs.FSProviderFromContext(ctx) == nil {
		ctx = datafs.ContextWithFSProvider(ctx, DefaultFSProvider)
	}

	// extract the rendering options from the config
	opts := optionsFromConfig(cfg)
	opts.Funcs = funcMap
	tr := newRenderer(opts)

	start := time.Now()

	// figure out how to name output files (only relevant if we're dealing with an InputDir)
	namer := chooseNamer(cfg, tr)

	// prepare to render templates (read them in, open output writers, etc)
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

type outputNamer interface {
	// Name the output file for the given input path
	Name(ctx context.Context, inPath string) (string, error)
}

type outputNamerFunc func(context.Context, string) (string, error)

func (f outputNamerFunc) Name(ctx context.Context, inPath string) (string, error) {
	return f(ctx, inPath)
}

func chooseNamer(cfg *Config, tr *renderer) outputNamer {
	if cfg.OutputMap == "" {
		return simpleNamer(cfg.OutputDir)
	}
	return mappingNamer(cfg.OutputMap, tr)
}

func simpleNamer(outDir string) outputNamer {
	return outputNamerFunc(func(_ context.Context, inPath string) (string, error) {
		outPath := filepath.Join(outDir, inPath)
		return filepath.Clean(outPath), nil
	})
}

func mappingNamer(outMap string, tr *renderer) outputNamer {
	return outputNamerFunc(func(ctx context.Context, inPath string) (string, error) {
		tcontext, err := createTmplContext(ctx, tr.tctxAliases, tr.sr)
		if err != nil {
			return "", err
		}

		// add '.in' to the template context and preserve the original context
		// in '.ctx'
		tctx := &tmplctx{}
		//nolint:gocritic
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
	})
}
