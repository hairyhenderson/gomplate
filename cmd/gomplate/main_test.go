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

func TestProcessIncludes(t *testing.T) {
	data := []struct {
		inc, exc, expected []string
	}{
		{nil, nil, []string{}},
		{[]string{}, []string{}, []string{}},
		{nil, []string{"*.foo"}, []string{"*.foo"}},
		{[]string{"*.bar"}, []string{"a*.bar"}, []string{"*", "!*.bar", "a*.bar"}},
		{[]string{"*.bar"}, nil, []string{"*", "!*.bar"}},
	}

	for _, d := range data {
		assert.EqualValues(t, d.expected, processIncludes(d.inc, d.exc))
	}
}
