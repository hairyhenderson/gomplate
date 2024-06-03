package cmd

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"

	"github.com/hairyhenderson/gomplate/v4"
	"github.com/hairyhenderson/gomplate/v4/env"
	"github.com/hairyhenderson/gomplate/v4/internal/datafs"
	"github.com/hairyhenderson/gomplate/v4/version"

	"github.com/spf13/cobra"
)

// postRunExec - if templating succeeds, the command following a '--' will be executed
func postRunExec(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	if len(args) > 0 {
		slog.DebugContext(ctx, "running post-exec command", "args", args)

		//nolint:govet
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		name := args[0]
		args = args[1:]

		c := exec.CommandContext(ctx, name, args...)
		c.Stdin = stdin
		c.Stderr = stderr
		c.Stdout = stdout

		// make sure all signals are propagated
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs)

		err := c.Start()
		if err != nil {
			return err
		}

		go func() {
			select {
			case sig := <-sigs:
				// Pass signals to the sub-process
				if c.Process != nil {
					_ = c.Process.Signal(sig)
				}
			case <-ctx.Done():
			}
		}()

		return c.Wait()
	}
	return nil
}

// optionalExecArgs - implements cobra.PositionalArgs. Allows extra args following
// a '--', but not otherwise.
func optionalExecArgs(cmd *cobra.Command, args []string) error {
	if cmd.ArgsLenAtDash() == 0 {
		return nil
	}
	return cobra.NoArgs(cmd, args)
}

// NewGomplateCmd -
func NewGomplateCmd(stderr io.Writer) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "gomplate",
		Short:   "Process text files with Go templates",
		Version: version.Version,
		RunE: func(cmd *cobra.Command, args []string) error {
			level := slog.LevelWarn
			if v, _ := cmd.Flags().GetBool("verbose"); v {
				level = slog.LevelDebug
			}
			initLogger(stderr, level)

			ctx := cmd.Context()

			cfg, err := loadConfig(ctx, cmd, args)
			if err != nil {
				return err
			}

			// get the post-exec reader now as this may modify cfg
			postExecReader := postExecInput(cfg)

			slog.DebugContext(ctx, fmt.Sprintf("starting %s", cmd.Name()))
			slog.DebugContext(ctx, fmt.Sprintf("config is:\n%v", cfg),
				slog.String("version", version.Version),
				slog.String("build", version.GitCommit),
			)

			// run the main command
			err = gomplate.Run(ctx, cfg)
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true

			slog.DebugContext(ctx, "completed rendering",
				slog.Int("templatesRendered", gomplate.Metrics.TemplatesProcessed),
				slog.Int("errors", gomplate.Metrics.Errors),
				slog.Duration("duration", gomplate.Metrics.TotalRenderDuration))

			if err != nil {
				return err
			}

			return postRunExec(ctx, cfg.PostExec, postExecReader, cmd.OutOrStdout(), cmd.ErrOrStderr())
		},
		Args: optionalExecArgs,
	}
	return rootCmd
}

// InitFlags - initialize the various flags and help strings on the command.
// Note that the defaults set here are ignored, and instead defaults from
// *config.Config's ApplyDefaults method are used instead. Changes here must be
// reflected there as well.
func InitFlags(command *cobra.Command) {
	command.Flags().SortFlags = false

	command.Flags().StringSliceP("datasource", "d", nil, "`datasource` in alias=URL form. Specify multiple times to add multiple sources.")
	command.Flags().StringSliceP("datasource-header", "H", nil, "HTTP `header` field in 'alias=Name: value' form to be provided on HTTP-based data sources. Multiples can be set.")

	command.Flags().StringSliceP("context", "c", nil, "pre-load a `datasource` into the context, in alias=URL form. Use the special alias `.` to set the root context.")

	command.Flags().StringSlice("plugin", nil, "plug in an external command as a function in name=path form. Can be specified multiple times")

	command.Flags().StringSliceP("file", "f", []string{"-"}, "Template `file` to process. Omit to use standard input, or use --in or --input-dir")
	command.Flags().StringP("in", "i", "", "Template `string` to process (alternative to --file and --input-dir)")
	command.Flags().String("input-dir", "", "`directory` which is examined recursively for templates (alternative to --file and --in)")

	command.Flags().StringSlice("exclude", []string{}, "glob of files to not parse")
	command.Flags().StringSlice("exclude-processing", []string{}, "glob of files to be copied without parsing")
	command.Flags().StringSlice("include", []string{}, "glob of files to parse")

	command.Flags().StringSliceP("out", "o", []string{"-"}, "output `file` name. Omit to use standard output.")
	command.Flags().StringSliceP("template", "t", []string{}, "Additional template file(s)")
	command.Flags().String("output-dir", ".", "`directory` to store the processed templates. Only used for --input-dir")
	command.Flags().String("output-map", "", "Template `string` to map the input file to an output path")
	command.Flags().String("chmod", "", "set the mode for output file(s). Omit to inherit from input file(s)")

	command.Flags().Bool("exec-pipe", false, "pipe the output to the post-run exec command")

	// these are only set for the help output - these defaults aren't actually used
	ldDefault := env.Getenv("GOMPLATE_LEFT_DELIM", "{{")
	rdDefault := env.Getenv("GOMPLATE_RIGHT_DELIM", "}}")
	command.Flags().String("left-delim", ldDefault, "override the default left-`delimiter` [$GOMPLATE_LEFT_DELIM]")
	command.Flags().String("right-delim", rdDefault, "override the default right-`delimiter` [$GOMPLATE_RIGHT_DELIM]")

	command.Flags().String("missing-key", "error", "Control the behavior during execution if a map is indexed with a key that is not present in the map. error (default) - return an error, zero - fallback to zero value, default/invalid - print <no value>")

	command.Flags().Bool("experimental", false, "enable experimental features [$GOMPLATE_EXPERIMENTAL]")

	command.Flags().BoolP("verbose", "V", false, "output extra information about what gomplate is doing")

	command.Flags().String("config", defaultConfigFile, "config file (overridden by commandline flags)")
}

// Main -
func Main(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	// inject default filesystem provider if it hasn't already been provided in
	// the context
	if datafs.FSProviderFromContext(ctx) == nil {
		ctx = datafs.ContextWithFSProvider(ctx, gomplate.DefaultFSProvider)
	}

	command := NewGomplateCmd(stderr)
	InitFlags(command)
	command.SetArgs(args)
	command.SetIn(stdin)
	command.SetOut(stdout)
	command.SetErr(stderr)

	err := command.ExecuteContext(ctx)
	if err != nil {
		slog.Error("", slog.Any("err", err))
	}
	return err
}
