package main

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestValidateOpts(t *testing.T) {
	err := validateOpts(parseFlags(), nil)
	assert.NoError(t, err)

	cmd := parseFlags("-i=foo", "-f", "bar")
	err = validateOpts(cmd, nil)
	assert.Error(t, err)

	cmd = parseFlags("-i=foo", "-o=bar", "-o=baz")
	err = validateOpts(cmd, nil)
	assert.Error(t, err)

	cmd = parseFlags("-i=foo", "--input-dir=baz")
	err = validateOpts(cmd, nil)
	assert.Error(t, err)

	cmd = parseFlags("--input-dir=foo", "-f=bar")
	err = validateOpts(cmd, nil)
	assert.Error(t, err)

	cmd = parseFlags("--output-dir=foo", "-o=bar")
	err = validateOpts(cmd, nil)
	assert.Error(t, err)

	cmd = parseFlags("--output-dir=foo")
	err = validateOpts(cmd, nil)
	assert.Error(t, err)

	cmd = parseFlags("--output-map", "bar")
	err = validateOpts(cmd, nil)
	assert.Error(t, err)

	cmd = parseFlags("-o", "foo", "--output-map", "bar")
	err = validateOpts(cmd, nil)
	assert.Error(t, err)

	cmd = parseFlags(
		"--input-dir", "in",
		"--output-dir", "foo",
		"--output-map", "bar",
	)
	err = validateOpts(cmd, nil)
	assert.Error(t, err)

	cmd = parseFlags(
		"--input-dir", "in",
		"--output-map", "bar",
	)
	err = validateOpts(cmd, nil)
	assert.NoError(t, err)
}

func parseFlags(flags ...string) *cobra.Command {
	cmd := &cobra.Command{}
	initFlags(cmd)
	err := cmd.ParseFlags(flags)
	if err != nil {
		panic(err)
	}
	return cmd
}
