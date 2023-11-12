package cmd

import (
	"context"
	"io"
	"os"
	"os/exec"
	"os/signal"

	"github.com/hairyhenderson/go-fsimpl/filefs"
	"github.com/hairyhenderson/gomplate/v3"
	"github.com/hairyhenderson/gomplate/v3/env"
	"github.com/hairyhenderson/gomplate/v3/version"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

// postRunExec - if templating succeeds, the command following a '--' will be executed
func postRunExec(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	if len(args) > 0 {
		log := zerolog.Ctx(ctx)
		log.Debug().Strs("args", args).Msg("running post-exec command")

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
func NewGomplateCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "gomplate",
		Short:   "Process text files with Go templates",
		Version: version.Version,
		RunE: func(cmd *cobra.Command, args []string) error {
			if v, _ := cmd.Flags().GetBool("verbose"); v {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}
			ctx := cmd.Context()
			log := zerolog.Ctx(ctx)

			cfg, err := loadConfig(cmd, args)
			if err != nil {
				return err
			}

			if cfg.Experimental {
				log.UpdateContext(func(c zerolog.Context) zerolog.Context {
					return c.Bool("experimental", true)
				})
				log.Info().Msg("experimental functions and features enabled!")

				ctx = gomplate.SetExperimental(ctx)
			}

			log.Debug().Msgf("starting %s", cmd.Name())
			log.Debug().
				Str("version", version.Version).
				Str("build", version.GitCommit).
				Msgf("config is:\n%v", cfg)

			err = gomplate.Run(ctx, cfg)
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true

			log.Debug().Int("templatesRendered", gomplate.Metrics.TemplatesProcessed).
				Int("errors", gomplate.Metrics.Errors).
				Dur("duration", gomplate.Metrics.TotalRenderDuration).
				Msg("completed rendering")

			if err != nil {
				return err
			}
			return postRunExec(ctx, cfg.PostExec, cfg.PostExecInput, cmd.OutOrStdout(), cmd.ErrOrStderr())
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

	command.Flags().Bool("experimental", false, "enable experimental features [$GOMPLATE_EXPERIMENTAL]")

	command.Flags().BoolP("verbose", "V", false, "output extra information about what gomplate is doing")

	command.Flags().String("config", defaultConfigFile, "config file (overridden by commandline flags)")
}

// Main -
func Main(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	ctx = initLogger(ctx, stderr)

	// inject a default filesystem provider for file:// URLs
	// TODO: expand this to support other schemes!
	if gomplate.FSProviderFromContext(ctx) == nil {
		// allow this to be overridden by tests
		ctx = gomplate.ContextWithFSProvider(ctx, filefs.FS)
	}

	command := NewGomplateCmd()
	InitFlags(command)
	command.SetArgs(args)
	command.SetIn(stdin)
	command.SetOut(stdout)
	command.SetErr(stderr)

	err := command.ExecuteContext(ctx)
	if err != nil {
		log := zerolog.Ctx(ctx)
		log.Error().Err(err).Send()
	}
	return err
}
