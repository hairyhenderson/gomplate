package gomplate

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigString(t *testing.T) {
	c := &Config{}

	expected := `input: -
output: -`
	assert.Equal(t, expected, c.String())

	c = &Config{
		LDelim:      "{{",
		RDelim:      "}}",
		Input:       "{{ foo }}",
		OutputFiles: []string{"-"},
		Templates:   []string{"foo=foo.t", "bar=bar.t"},
	}
	expected = `input: <arg>
output: -
templates: foo=foo.t, bar=bar.t`

	assert.Equal(t, expected, c.String())
}

func TestGetMode(t *testing.T) {
	c := &Config{}
	m, o, err := c.getMode()
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0), m)
	assert.False(t, o)

	c = &Config{OutMode: "755"}
	m, o, err = c.getMode()
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0755), m)
	assert.True(t, o)

	c = &Config{OutMode: "0755"}
	m, o, err = c.getMode()
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0755), m)
	assert.True(t, o)

	c = &Config{OutMode: "foo"}
	_, _, err = c.getMode()
	assert.Error(t, err)
}
