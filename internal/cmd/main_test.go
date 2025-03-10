package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOptionalExecArgs(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetArgs(nil)
	cmd.ParseFlags(nil)

	err := optionalExecArgs(cmd, nil)
	require.NoError(t, err)

	cmd = &cobra.Command{}
	cmd.SetArgs(nil)
	cmd.ParseFlags(nil)

	err = optionalExecArgs(cmd, []string{"bogus"})
	require.Error(t, err)

	cmd = &cobra.Command{}
	cmd.SetArgs(nil)
	cmd.ParseFlags([]string{"--", "foo"})

	err = optionalExecArgs(cmd, []string{})
	require.NoError(t, err)

	cmd = &cobra.Command{}
	cmd.SetArgs(nil)
	cmd.ParseFlags([]string{"--"})

	err = optionalExecArgs(cmd, []string{"foo"})
	require.NoError(t, err)
}

func TestRunMain(t *testing.T) {
	ctx := t.Context()

	err := Main(ctx, []string{"-h"}, nil, nil, nil)
	require.NoError(t, err)

	err = Main(ctx, []string{"--bogus"}, nil, nil, nil)
	require.Error(t, err)

	stdin := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	err = Main(ctx, []string{"-i", "hello"}, stdin, stdout, stderr)
	require.NoError(t, err)
	assert.Equal(t, "hello", stdout.String())
}

func TestPostRunExec(t *testing.T) {
	ctx := t.Context()

	out := &bytes.Buffer{}
	err := postRunExec(ctx, []string{"cat"}, strings.NewReader("hello world"), out, out)
	require.NoError(t, err)
	assert.Equal(t, "hello world", out.String())
}
