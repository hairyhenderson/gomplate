package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/hairyhenderson/gomplate/env"
	"github.com/hairyhenderson/gomplate/version"
	"github.com/spf13/cobra"
)

// GomplateOpts -
type GomplateOpts struct {
	version           bool
	dataSources       []string
	dataSourceHeaders []string
	lDelim            string
	rDelim            string

	input       string
	inputFiles  []string
	inputDir    string
	outputFiles []string
	outputDir   string
	excludeGlob string
}

var opts GomplateOpts

func validateOpts(cmd *cobra.Command, args []string) error {
	if cmd.Flag("in").Changed && cmd.Flag("file").Changed {
		return errors.New("--in and --file may not be used together")
	}

	if len(opts.inputFiles) != len(opts.outputFiles) {
		return fmt.Errorf("Must provide same number of --out (%d) as --file (%d) options", len(opts.outputFiles), len(opts.inputFiles))
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
	// fmt.Printf("%s version %s, build %s\n", name, version.Version, version.GitCommit)
	fmt.Printf("%s version %s\n", name, version.Version)
}

func newGomplateCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "gomplate",
		Short:   "Process text files with Go templates",
		PreRunE: validateOpts,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.version {
				printVersion(cmd.Name())
				return nil
			}
			return runTemplate(&opts)
		},
		Args: cobra.NoArgs,
	}
	return rootCmd
}

func initFlags(command *cobra.Command) {
	command.Flags().BoolVarP(&opts.version, "version", "v", false, "print the version")

	command.Flags().StringArrayVarP(&opts.inputFiles, "file", "f", []string{"-"}, "Template `file` to process. Omit to use standard input, or use --in or --input-dir")
	command.Flags().StringVarP(&opts.input, "in", "i", "", "Template `string` to process (alternative to --file and --input-dir)")
	command.Flags().StringVar(&opts.inputDir, "input-dir", "", "`directory` which is examined recursively for templates (alternative to --file and --in)")
	command.Flags().StringVar(&opts.excludeGlob, "exclude", "", "glob of files to not parse")
	command.Flags().StringArrayVarP(&opts.outputFiles, "out", "o", []string{"-"}, "output `file` name. Omit to use standard output.")
	command.Flags().StringVar(&opts.outputDir, "output-dir", ".", "`directory` to store the processed templates. Only used for --input-dir")

	command.Flags().StringArrayVarP(&opts.dataSources, "datasource", "d", nil, "`datasource` in alias=URL form. Specify multiple times to add multiple sources.")
	command.Flags().StringArrayVarP(&opts.dataSourceHeaders, "datasource-header", "H", nil, "HTTP `header` field in 'alias=Name: value' form to be provided on HTTP-based data sources. Multiples can be set.")

	ldDefault := env.Getenv("GOMPLATE_LEFT_DELIM", "{{")
	rdDefault := env.Getenv("GOMPLATE_RIGHT_DELIM", "}}")
	command.Flags().StringVar(&opts.lDelim, "left-delim", ldDefault, "override the default left-`delimiter` [$GOMPLATE_LEFT_DELIM]")
	command.Flags().StringVar(&opts.rDelim, "right-delim", rdDefault, "override the default right-`delimiter` [$GOMPLATE_RIGHT_DELIM]")
}

func main() {
	command := newGomplateCmd()
	initFlags(command)
	if err := command.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
