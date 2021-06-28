package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/hairyhenderson/gomplate/v3/conv"
	"github.com/hairyhenderson/gomplate/v3/env"
	"github.com/hairyhenderson/gomplate/v3/internal/config"

	"github.com/rs/zerolog"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

const (
	defaultConfigFile = ".gomplate.yaml"
)

var fs = afero.NewOsFs()

// loadConfig is intended to be called before command execution. It:
// - creates a config.Config from the cobra flags
// - creates a config.Config from the config file (if present)
// - merges the two (flags take precedence)
// - validates the final config
func loadConfig(cmd *cobra.Command, args []string) (*config.Config, error) {
	ctx := cmd.Context()
	flagConfig, err := cobraConfig(cmd, args)
	if err != nil {
		return nil, err
	}

	cfg, err := readConfigFile(cmd)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		cfg = flagConfig
	} else {
		cfg = cfg.MergeFrom(flagConfig)
	}

	cfg, err = applyEnvVars(ctx, cfg)
	if err != nil {
		return nil, err
	}

	cfg.Stdin = cmd.InOrStdin()
	cfg.Stdout = cmd.OutOrStdout()
	cfg.Stderr = cmd.ErrOrStderr()

	// reset defaults before validation
	cfg.ApplyDefaults()

	err = cfg.Validate()
	if err != nil {
		return nil, fmt.Errorf("failed to validate merged config: %w\n%+v", err, cfg)
	}
	return cfg, nil
}

func pickConfigFile(cmd *cobra.Command) (cfgFile string, required bool) {
	cfgFile = defaultConfigFile
	if c := env.Getenv("GOMPLATE_CONFIG"); c != "" {
		cfgFile = c
		required = true
	}
	if cmd.Flags().Changed("config") && cmd.Flag("config").Value.String() != "" {
		// Use config file from the flag if specified
		cfgFile = cmd.Flag("config").Value.String()
		required = true
	}
	return cfgFile, required
}

func readConfigFile(cmd *cobra.Command) (cfg *config.Config, err error) {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	log := zerolog.Ctx(ctx)

	cfgFile, configRequired := pickConfigFile(cmd)

	f, err := fs.Open(cfgFile)
	if err != nil {
		if configRequired {
			return cfg, fmt.Errorf("config file requested, but couldn't be opened: %w", err)
		}
		return nil, nil
	}

	cfg, err = config.Parse(f)
	if err != nil && configRequired {
		return cfg, fmt.Errorf("config file requested, but couldn't be parsed: %w", err)
	}

	log.Debug().Str("cfgFile", cfgFile).Msg("using config file")

	return cfg, err
}

// cobraConfig - initialize a config from the commandline options
func cobraConfig(cmd *cobra.Command, args []string) (cfg *config.Config, err error) {
	cfg = &config.Config{}
	cfg.InputFiles, err = getStringSlice(cmd, "file")
	if err != nil {
		return nil, err
	}
	cfg.Input, err = getString(cmd, "in")
	if err != nil {
		return nil, err
	}
	cfg.InputDir, err = getString(cmd, "input-dir")
	if err != nil {
		return nil, err
	}

	cfg.ExcludeGlob, err = getStringSlice(cmd, "exclude")
	if err != nil {
		return nil, err
	}
	includesFlag, err := getStringSlice(cmd, "include")
	if err != nil {
		return nil, err
	}
	// support --include
	cfg.ExcludeGlob = processIncludes(includesFlag, cfg.ExcludeGlob)

	cfg.OutputFiles, err = getStringSlice(cmd, "out")
	if err != nil {
		return nil, err
	}
	cfg.OutputDir, err = getString(cmd, "output-dir")
	if err != nil {
		return nil, err
	}
	cfg.OutputMap, err = getString(cmd, "output-map")
	if err != nil {
		return nil, err
	}
	cfg.OutMode, err = getString(cmd, "chmod")
	if err != nil {
		return nil, err
	}

	if len(args) > 0 {
		cfg.PostExec = args
	}

	cfg.ExecPipe, err = getBool(cmd, "exec-pipe")
	if err != nil {
		return nil, err
	}
	cfg.Experimental, err = getBool(cmd, "experimental")
	if err != nil {
		return nil, err
	}

	cfg.LDelim, err = getString(cmd, "left-delim")
	if err != nil {
		return nil, err
	}
	cfg.RDelim, err = getString(cmd, "right-delim")
	if err != nil {
		return nil, err
	}

	ds, err := getStringSlice(cmd, "datasource")
	if err != nil {
		return nil, err
	}

	cx, err := getStringSlice(cmd, "context")
	if err != nil {
		return nil, err
	}

	err = cfg.ParseDataSourceFlags(ds, cx)
	if err != nil {
		return nil, err
	}

	ts, err := getStringSlice(cmd, "template")
	if err != nil {
		return nil, err
	}

	err = cfg.ParseTemplateFlags(ts)
	if err != nil {
		return nil, err
	}

	hdr, err := getStringSlice(cmd, "datasource-header")
	if err != nil {
		return nil, err
	}

	err = cfg.ParseHeaderFlags(hdr)
	if err != nil {
		return nil, err
	}

	pl, err := getStringSlice(cmd, "plugin")
	if err != nil {
		return nil, err
	}

	err = cfg.ParsePluginFlags(pl)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func getStringSlice(cmd *cobra.Command, flag string) (s []string, err error) {
	if cmd.Flag(flag) != nil && cmd.Flag(flag).Changed {
		s, err = cmd.Flags().GetStringSlice(flag)
	}
	return s, err
}

func getString(cmd *cobra.Command, flag string) (s string, err error) {
	if cmd.Flag(flag) != nil && cmd.Flag(flag).Changed {
		s, err = cmd.Flags().GetString(flag)
	}
	return s, err
}

func getBool(cmd *cobra.Command, flag string) (b bool, err error) {
	if cmd.Flag(flag) != nil && cmd.Flag(flag).Changed {
		b, err = cmd.Flags().GetBool(flag)
	}
	return b, err
}

// process --include flags - these are analogous to specifying --exclude '*',
// then the inverse of the --include options.
func processIncludes(includes, excludes []string) []string {
	if len(includes) == 0 && len(excludes) == 0 {
		return nil
	}

	out := []string{}
	// if any --includes are set, we start by excluding everything
	if len(includes) > 0 {
		out = make([]string, 1+len(includes))
		out[0] = "*"
	}
	for i, include := range includes {
		// includes are just the opposite of an exclude
		out[i+1] = "!" + include
	}
	out = append(out, excludes...)
	return out
}

func applyEnvVars(ctx context.Context, cfg *config.Config) (*config.Config, error) {
	if to := env.Getenv("GOMPLATE_PLUGIN_TIMEOUT"); cfg.PluginTimeout == 0 && to != "" {
		t, err := time.ParseDuration(to)
		if err != nil {
			return nil, fmt.Errorf("GOMPLATE_PLUGIN_TIMEOUT set to invalid value %q: %w", to, err)
		}
		cfg.PluginTimeout = t
	}

	if !cfg.SuppressEmpty && conv.ToBool(env.Getenv("GOMPLATE_SUPPRESS_EMPTY", "false")) {
		cfg.SuppressEmpty = true
	}

	if !cfg.Experimental && conv.ToBool(env.Getenv("GOMPLATE_EXPERIMENTAL", "false")) {
		cfg.Experimental = true
	}

	if cfg.LDelim == "" {
		cfg.LDelim = env.Getenv("GOMPLATE_LEFT_DELIM")
	}
	if cfg.RDelim == "" {
		cfg.RDelim = env.Getenv("GOMPLATE_RIGHT_DELIM")
	}

	return cfg, nil
}
