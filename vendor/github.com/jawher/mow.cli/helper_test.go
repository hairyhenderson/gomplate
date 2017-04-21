package cli

import (
	"testing"

	"bytes"

	"io/ioutil"

	"github.com/stretchr/testify/require"
)

func exitShouldBeCalledWith(t *testing.T, wantedExitCode int, called *bool) func() {
	oldExiter := exiter
	exiter = func(code int) {
		require.Equal(t, wantedExitCode, code, "unwanted exit code")
		*called = true
	}
	return func() { exiter = oldExiter }
}

func exitShouldNotCalled(t *testing.T) func() {
	oldExiter := exiter
	exiter = func(code int) {
		t.Errorf("exit should not have been called")
	}
	return func() { exiter = oldExiter }
}

func suppressOutput() func() {
	return captureAndRestoreOutput(nil, nil)
}

func captureAndRestoreOutput(out, err *string) func() {
	oldStdOut := stdOut
	oldStdErr := stdErr

	if out == nil {
		stdOut = ioutil.Discard
	} else {
		stdOut = trapWriter(out)
	}
	if err == nil {
		stdErr = ioutil.Discard
	} else {
		stdErr = trapWriter(err)
	}

	return func() {
		stdOut = oldStdOut
		stdErr = oldStdErr
	}
}

func trapWriter(writeTo *string) *writerTrap {
	return &writerTrap{
		buffer:  bytes.NewBuffer(nil),
		writeTo: writeTo,
	}
}

type writerTrap struct {
	buffer  *bytes.Buffer
	writeTo *string
}

func (w *writerTrap) Write(p []byte) (n int, err error) {
	n, err = w.buffer.Write(p)
	if err == nil {
		*(w.writeTo) = w.buffer.String()
	}
	return
}
