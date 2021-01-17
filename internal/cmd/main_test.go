package cmd

import (
	"bytes"
	"context"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestOptionalExecArgs(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetArgs(nil)
	cmd.ParseFlags(nil)

	err := optionalExecArgs(cmd, nil)
	assert.NoError(t, err)

	cmd = &cobra.Command{}
	cmd.SetArgs(nil)
	cmd.ParseFlags(nil)

	err = optionalExecArgs(cmd, []string{"bogus"})
	assert.Error(t, err)

	cmd = &cobra.Command{}
	cmd.SetArgs(nil)
	cmd.ParseFlags([]string{"--", "foo"})

	err = optionalExecArgs(cmd, []string{})
	assert.NoError(t, err)

	cmd = &cobra.Command{}
	cmd.SetArgs(nil)
	cmd.ParseFlags([]string{"--"})

	err = optionalExecArgs(cmd, []string{"foo"})
	assert.NoError(t, err)
}

func TestRunMain(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := Main(ctx, []string{"-h"}, nil, nil, nil)
	assert.NoError(t, err)

	err = Main(ctx, []string{"--bogus"}, nil, nil, nil)
	assert.Error(t, err)

	stdin := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	err = Main(ctx, []string{"-i", "hello"}, stdin, stdout, stderr)
	assert.NoError(t, err)
	assert.Equal(t, "hello", stdout.String())
}
