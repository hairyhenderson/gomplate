package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/hairyhenderson/gomplate/v4"
	"github.com/hairyhenderson/gomplate/v4/conv"
	"github.com/hairyhenderson/gomplate/v4/env"
	"github.com/hairyhenderson/gomplate/v4/internal/datafs"
	"github.com/hairyhenderson/gomplate/v4/internal/urlhelpers"

	"github.com/spf13/cobra"
)

const (
	defaultConfigFile = ".gomplate.yaml"
)

// loadConfig is intended to be called before command execution. It:
// - creates a gomplate.Config from the cobra flags
// - creates a gomplate.Config from the config file (if present)
// - merges the two (flags take precedence)
func loadConfig(ctx context.Context, cmd *cobra.Command, args []string) (*gomplate.Config, error) {
	flagConfig, err := cobraConfig(cmd, args)
	if err != nil {
		return nil, err
	}

	cfg, err := readConfigFile(ctx, cmd)
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

	return cfg, nil
}

func pickConfigFile(cmd *cobra.Command) (cfgFile string, required, skip bool) {
	cfgFile = defaultConfigFile
	if c, found := env.LookupEnv("GOMPLATE_CONFIG"); found {
		cfgFile = c
		if cfgFile == "" {
			skip = true
		} else {
			required = true
		}
	}
	if cmd.Flags().Changed("config") {
		// Use config file from the flag if specified
		cfgFile = cmd.Flag("config").Value.String()
		if cfgFile == "" {
			skip = true
		} else {
			required = true
		}
	}
	return cfgFile, required, skip
}

func readConfigFile(ctx context.Context, cmd *cobra.Command) (*gomplate.Config, error) {
	cfgFile, configRequired, skip := pickConfigFile(cmd)
	if skip {
		// --config was specified with an empty value
		return nil, nil
	}

	// we only support loading configs from the local filesystem for now
	fsys, err := datafs.FSysForPath(ctx, cfgFile)
	if err != nil {
		return nil, fmt.Errorf("fsys for path %v: %w", cfgFile, err)
	}

	f, err := fsys.Open(cfgFile)
	if err != nil {
		if configRequired {
			return nil, fmt.Errorf("config file requested, but couldn't be opened: %w", err)
		}
		return nil, nil
	}

	cfg, err := gomplate.Parse(f)
	if err != nil {
		return nil, fmt.Errorf("parsing config file %q: %w", cfgFile, err)
	}

	slog.DebugContext(ctx, "using config file", "cfgFile", cfgFile)

	return cfg, nil
}

// cobraConfig - initialize a config from the commandline options
func cobraConfig(cmd *cobra.Command, args []string) (cfg *gomplate.Config, err error) {
	cfg = &gomplate.Config{}
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
	cfg.ExcludeProcessingGlob, err = getStringSlice(cmd, "exclude-processing")
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

	cfg.MissingKey, err = getString(cmd, "missing-key")
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
	ts, err := getStringSlice(cmd, "template")
	if err != nil {
		return nil, err
	}
	hdr, err := getStringSlice(cmd, "datasource-header")
	if err != nil {
		return nil, err
	}
	err = ParseDataSourceFlags(cfg, ds, cx, ts, hdr)
	if err != nil {
		return nil, err
	}

	pl, err := getStringSlice(cmd, "plugin")
	if err != nil {
		return nil, err
	}
	err = ParsePluginFlags(cfg, pl)
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

func applyEnvVars(_ context.Context, cfg *gomplate.Config) (*gomplate.Config, error) {
	if to := env.Getenv("GOMPLATE_PLUGIN_TIMEOUT"); cfg.PluginTimeout == 0 && to != "" {
		t, err := time.ParseDuration(to)
		if err != nil {
			return nil, fmt.Errorf("GOMPLATE_PLUGIN_TIMEOUT set to invalid value %q: %w", to, err)
		}
		cfg.PluginTimeout = t
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

// postExecInput - return the input to be used after the post-exec command. The
// input config may be modified if ExecPipe is set (OutputFiles is set to "-"),
// and Stdout is redirected to a pipe.
func postExecInput(cfg *gomplate.Config) io.Reader {
	if cfg.ExecPipe {
		pipe := &bytes.Buffer{}
		cfg.OutputFiles = []string{"-"}

		// --exec-pipe redirects standard out to the out pipe
		cfg.Stdout = pipe

		return pipe
	}

	if cfg.Stdin != nil {
		return cfg.Stdin
	}

	return os.Stdin
}

// ParsePluginFlags - sets the Plugins field from the
// key=value format flags as provided at the command-line
func ParsePluginFlags(c *gomplate.Config, plugins []string) error {
	for _, plugin := range plugins {
		parts := strings.SplitN(plugin, "=", 2)
		if len(parts) < 2 {
			return fmt.Errorf("plugin requires both name and path")
		}
		if c.Plugins == nil {
			c.Plugins = map[string]gomplate.PluginConfig{}
		}
		c.Plugins[parts[0]] = gomplate.PluginConfig{Cmd: parts[1]}
	}
	return nil
}

// ParseDataSourceFlags - sets DataSources, Context, and Templates fields from
// the key=value format flags as provided at the command-line
// Unreferenced headers will be set in c.ExtraHeaders
func ParseDataSourceFlags(c *gomplate.Config, datasources, contexts, templates, headers []string) error {
	err := parseResources(c, datasources, contexts, templates)
	if err != nil {
		return err
	}

	hdrs, err := parseHeaderArgs(headers)
	if err != nil {
		return err
	}

	for k, v := range hdrs {
		if d, ok := c.Context[k]; ok {
			d.Header = v
			c.Context[k] = d
			delete(hdrs, k)
		}
		if d, ok := c.DataSources[k]; ok {
			d.Header = v
			c.DataSources[k] = d
			delete(hdrs, k)
		}
		if t, ok := c.Templates[k]; ok {
			t.Header = v
			c.Templates[k] = t
			delete(hdrs, k)
		}
	}
	if len(hdrs) > 0 {
		c.ExtraHeaders = hdrs
	}
	return nil
}

func parseResources(c *gomplate.Config, datasources, contexts, templates []string) error {
	for _, d := range datasources {
		k, ds, err := parseDatasourceArg(d)
		if err != nil {
			return err
		}
		if c.DataSources == nil {
			c.DataSources = map[string]gomplate.DataSource{}
		}
		c.DataSources[k] = ds
	}
	for _, d := range contexts {
		k, ds, err := parseDatasourceArg(d)
		if err != nil {
			return err
		}
		if c.Context == nil {
			c.Context = map[string]gomplate.DataSource{}
		}
		c.Context[k] = ds
	}
	for _, t := range templates {
		k, ds, err := parseTemplateArg(t)
		if err != nil {
			return err
		}
		if c.Templates == nil {
			c.Templates = map[string]gomplate.DataSource{}
		}
		c.Templates[k] = ds
	}

	return nil
}

func parseDatasourceArg(value string) (alias string, ds gomplate.DataSource, err error) {
	alias, u, _ := strings.Cut(value, "=")
	if u == "" {
		u = alias
		alias, _, _ = strings.Cut(value, ".")
		if path.Base(u) != u {
			err = fmt.Errorf("invalid argument (%s): must provide an alias with files not in working directory", value)
			return alias, ds, err
		}
	}

	ds.URL, err = urlhelpers.ParseSourceURL(u)

	return alias, ds, err
}

func parseHeaderArgs(headerArgs []string) (map[string]http.Header, error) {
	headers := make(map[string]http.Header)
	for _, v := range headerArgs {
		ds, name, value, err := splitHeaderArg(v)
		if err != nil {
			return nil, err
		}
		if _, ok := headers[ds]; !ok {
			headers[ds] = make(http.Header)
		}
		headers[ds][name] = append(headers[ds][name], strings.TrimSpace(value))
	}
	return headers, nil
}

func splitHeaderArg(arg string) (datasourceAlias, name, value string, err error) {
	parts := strings.SplitN(arg, "=", 2)
	if len(parts) != 2 {
		err = fmt.Errorf("invalid datasource-header option '%s'", arg)
		return "", "", "", err
	}
	datasourceAlias = parts[0]
	name, value, err = splitHeader(parts[1])
	return datasourceAlias, name, value, err
}

func splitHeader(header string) (name, value string, err error) {
	parts := strings.SplitN(header, ":", 2)
	if len(parts) != 2 {
		err = fmt.Errorf("invalid HTTP Header format '%s'", header)
		return "", "", err
	}
	name = http.CanonicalHeaderKey(parts[0])
	value = parts[1]
	return name, value, nil
}

func parseTemplateArg(value string) (alias string, ds gomplate.DataSource, err error) {
	alias, u, _ := strings.Cut(value, "=")
	if u == "" {
		u = alias
	}

	ds.URL, err = urlhelpers.ParseSourceURL(u)

	return alias, ds, err
}
