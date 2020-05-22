/*
The gomplate command

*/
package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"

	"github.com/hairyhenderson/gomplate/v3"
	"github.com/hairyhenderson/gomplate/v3/env"
	"github.com/hairyhenderson/gomplate/v3/internal/config"
	"github.com/hairyhenderson/gomplate/v3/version"

	"github.com/rs/zerolog"

	"github.com/spf13/cobra"
)

// postRunExec - if templating succeeds, the command following a '--' will be executed
func postRunExec(ctx context.Context, cfg *config.Config) error {
	args := cfg.PostExec
	if len(args) > 0 {
		log := zerolog.Ctx(ctx)
		log.Debug().Strs("args", args).Msg("running post-exec command")

		name := args[0]
		args = args[1:]
		// nolint: gosec
		c := exec.CommandContext(ctx, name, args...)
		c.Stdin = cfg.PostExecInput
		c.Stderr = os.Stderr
		c.Stdout = os.Stdout

		// make sure all signals are propagated
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs)
		go func() {
			// Pass signals to the sub-process
			sig := <-sigs
			if c.Process != nil {
				// nolint: gosec
				_ = c.Process.Signal(sig)
			}
		}()

		return c.Run()
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

func newGomplateCmd() *cobra.Command {
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

			log.Debug().Msgf("starting %s", cmd.Name())
			log.Debug().
				Str("version", version.Version).
				Str("build", version.GitCommit).
				Msgf("config is:\n%v", cfg)

			err = gomplate.Run(ctx, cfg)
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true

			fmt.Fprintf(os.Stderr, "\n")
			log.Debug().Int("templatesRendered", gomplate.Metrics.TemplatesProcessed).
				Int("errors", gomplate.Metrics.Errors).
				Dur("duration", gomplate.Metrics.TotalRenderDuration).
				Msg("completed rendering")

			if err != nil {
				return err
			}
			return postRunExec(ctx, cfg)
		},
		Args: optionalExecArgs,
	}
	return rootCmd
}

func initFlags(command *cobra.Command) {
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

	ldDefault := env.Getenv("GOMPLATE_LEFT_DELIM", "{{")
	rdDefault := env.Getenv("GOMPLATE_RIGHT_DELIM", "}}")
	command.Flags().String("left-delim", ldDefault, "override the default left-`delimiter` [$GOMPLATE_LEFT_DELIM]")
	command.Flags().String("right-delim", rdDefault, "override the default right-`delimiter` [$GOMPLATE_RIGHT_DELIM]")

	command.Flags().BoolP("verbose", "V", false, "output extra information about what gomplate is doing")

	command.Flags().String("config", defaultConfigFile, "config file (overridden by commandline flags)")
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = initLogger(ctx)

	command := newGomplateCmd()
	initFlags(command)
	if err := command.ExecuteContext(ctx); err != nil {
		log := zerolog.Ctx(ctx)
		log.Fatal().Err(err).Send()
	}
}
