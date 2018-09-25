package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/hairyhenderson/gomplate"
	"github.com/hairyhenderson/gomplate/env"
	"github.com/hairyhenderson/gomplate/version"
	"github.com/spf13/cobra"
)

var (
	printVer bool
	verbose  bool
	opts     gomplate.Config
)

func validateOpts(cmd *cobra.Command, args []string) error {
	if cmd.Flag("in").Changed && cmd.Flag("file").Changed {
		return errors.New("--in and --file may not be used together")
	}

	if len(opts.InputFiles) != len(opts.OutputFiles) {
		return fmt.Errorf("Must provide same number of --out (%d) as --file (%d) options", len(opts.OutputFiles), len(opts.InputFiles))
	}

	if cmd.Flag("input-dir").Changed && (cmd.Flag("in").Changed || cmd.Flag("file").Changed) {
		return errors.New("--input-dir can not be used together with --in or --file")
	}

	if cmd.Flag("output-dir").Changed {
		if cmd.Flag("out").Changed {
			return errors.New("--output-dir can not be used together with --out")
		}
		if !cmd.Flag("input-dir").Changed {
			return errors.New("--input-dir must be set when --output-dir is set")
		}
	}
	return nil
}

func printVersion(name string) {
	fmt.Printf("%s version %s\n", name, version.Version)
}

// postRunExec - if templating succeeds, the command following a '--' will be executed
func postRunExec(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		name := args[0]
		args = args[1:]
		// nolint: gosec
		c := exec.Command(name, args...)
		c.Stdin = os.Stdin
		c.Stderr = os.Stderr
		c.Stdout = os.Stdout
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
		PreRunE: validateOpts,
		RunE: func(cmd *cobra.Command, args []string) error {
			if printVer {
				printVersion(cmd.Name())
				return nil
			}
			if verbose {
				// nolint: errcheck
				fmt.Fprintf(os.Stderr, "%s version %s, build %s (%v)\nconfig is:\n%s\n\n",
					cmd.Name(), version.Version, version.GitCommit, version.BuildDate,
					&opts)
			}

			err := gomplate.RunTemplates(&opts)
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true
			if verbose {
				// nolint: errcheck
				fmt.Fprintf(os.Stderr, "rendered %d template(s) with %d error(s) in %v\n",
					gomplate.Metrics.TemplatesProcessed, gomplate.Metrics.Errors, gomplate.Metrics.TotalRenderDuration)
			}
			return err
		},
		PostRunE: postRunExec,
		Args:     optionalExecArgs,
	}
	return rootCmd
}

func initFlags(command *cobra.Command) {
	command.Flags().SortFlags = false

	command.Flags().StringArrayVarP(&opts.DataSources, "datasource", "d", nil, "`datasource` in alias=URL form. Specify multiple times to add multiple sources.")
	command.Flags().StringArrayVarP(&opts.DataSourceHeaders, "datasource-header", "H", nil, "HTTP `header` field in 'alias=Name: value' form to be provided on HTTP-based data sources. Multiples can be set.")

	command.Flags().StringArrayVarP(&opts.InputFiles, "file", "f", []string{"-"}, "Template `file` to process. Omit to use standard input, or use --in or --input-dir")
	command.Flags().StringVarP(&opts.Input, "in", "i", "", "Template `string` to process (alternative to --file and --input-dir)")
	command.Flags().StringVar(&opts.InputDir, "input-dir", "", "`directory` which is examined recursively for templates (alternative to --file and --in)")

	command.Flags().StringArrayVar(&opts.ExcludeGlob, "exclude", []string{}, "glob of files to not parse")

	command.Flags().StringArrayVarP(&opts.OutputFiles, "out", "o", []string{"-"}, "output `file` name. Omit to use standard output.")
	command.Flags().StringArrayVarP(&opts.Templates, "template", "t", []string{}, "Additional template file(s)")
	command.Flags().StringVar(&opts.OutputDir, "output-dir", ".", "`directory` to store the processed templates. Only used for --input-dir")
	command.Flags().StringVar(&opts.OutMode, "chmod", "", "set the mode for output file(s). Omit to inherit from input file(s)")

	ldDefault := env.Getenv("GOMPLATE_LEFT_DELIM", "{{")
	rdDefault := env.Getenv("GOMPLATE_RIGHT_DELIM", "}}")
	command.Flags().StringVar(&opts.LDelim, "left-delim", ldDefault, "override the default left-`delimiter` [$GOMPLATE_LEFT_DELIM]")
	command.Flags().StringVar(&opts.RDelim, "right-delim", rdDefault, "override the default right-`delimiter` [$GOMPLATE_RIGHT_DELIM]")

	command.Flags().BoolVarP(&verbose, "verbose", "V", false, "output extra information about what gomplate is doing")

	command.Flags().BoolVarP(&printVer, "version", "v", false, "print the version")
}

func main() {
	command := newGomplateCmd()
	initFlags(command)
	if err := command.Execute(); err != nil {
		// nolint: errcheck
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
