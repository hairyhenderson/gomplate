package iohelpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMimeAlias(t *testing.T) {
	t.Parallel()
	data := []struct {
		in, out string
	}{
		{CSVMimetype, CSVMimetype},
		{YAMLMimetype, YAMLMimetype},
		{"application/x-yaml", YAMLMimetype},
	}

	for _, d := range data {
		assert.Equal(t, d.out, MimeAlias(d.in))
	}
}
