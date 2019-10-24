package main

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestValidateOpts(t *testing.T) {
	err := validateOpts(parseFlags())
	assert.NoError(t, err)

	err = validateOpts(parseFlags("-i=foo", "-f", "bar"))
	assert.Error(t, err)

	err = validateOpts(parseFlags("-i=foo", "-o=bar", "-o=baz"))
	assert.Error(t, err)

	err = validateOpts(parseFlags("-i=foo", "--input-dir=baz"))
	assert.Error(t, err)

	err = validateOpts(parseFlags("--input-dir=foo", "-f=bar"))
	assert.Error(t, err)

	err = validateOpts(parseFlags("--output-dir=foo", "-o=bar"))
	assert.Error(t, err)

	err = validateOpts(parseFlags("--output-dir=foo"))
	assert.Error(t, err)

	err = validateOpts(parseFlags("--output-map", "bar"))
	assert.Error(t, err)

	err = validateOpts(parseFlags("-o", "foo", "--output-map", "bar"))
	assert.Error(t, err)

	err = validateOpts(parseFlags(
		"--input-dir", "in",
		"--output-dir", "foo",
		"--output-map", "bar",
	))
	assert.Error(t, err)

	err = validateOpts(parseFlags(
		"--input-dir", "in",
		"--output-map", "bar",
	))
	assert.NoError(t, err)
}

func parseFlags(flags ...string) (cmd *cobra.Command, args []string) {
	cmd = &cobra.Command{}
	initFlags(cmd)
	err := cmd.ParseFlags(flags)
	if err != nil {
		panic(err)
	}
	return cmd, cmd.Flags().Args()
}
