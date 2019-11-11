package gomplate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigString(t *testing.T) {
	c := &Config{}

	expected := `input: -
output: -`
	assert.Equal(t, expected, c.String())

	c = &Config{
		LDelim:      "L",
		RDelim:      "R",
		Input:       "foo",
		OutputFiles: []string{"-"},
		Templates:   []string{"foo=foo.t", "bar=bar.t"},
	}
	expected = `input: <arg>
output: -
left_delim: L
right_delim: R
templates: foo=foo.t, bar=bar.t`

	assert.Equal(t, expected, c.String())

	c = &Config{
		InputDir:  "in/",
		OutputDir: "out/",
	}
	expected = `input: in/
output: out/`

	assert.Equal(t, expected, c.String())

	c = &Config{
		InputDir:  "in/",
		OutputMap: "{{ .in }}",
	}
	expected = `input: in/
output: {{ .in }}`

	assert.Equal(t, expected, c.String())
}
