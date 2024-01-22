package data

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMimeAlias(t *testing.T) {
	t.Parallel()
	data := []struct {
		in, out string
	}{
		{csvMimetype, csvMimetype},
		{yamlMimetype, yamlMimetype},
		{"application/x-yaml", yamlMimetype},
	}

	for _, d := range data {
		assert.Equal(t, d.out, mimeAlias(d.in))
	}
}
