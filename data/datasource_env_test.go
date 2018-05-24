package data

import (
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func mustParseURL(in string) *url.URL {
	u, _ := url.Parse(in)
	return u
}

func TestReadEnv(t *testing.T) {
	content := []byte(`hello world`)
	os.Setenv("HELLO_WORLD", "hello world")
	defer os.Unsetenv("HELLO_WORLD")
	os.Setenv("HELLO_UNIVERSE", "hello universe")
	defer os.Unsetenv("HELLO_UNIVERSE")

	source, _ := NewSource("foo", mustParseURL("env:HELLO_WORLD"))

	actual, err := readEnv(source)
	assert.NoError(t, err)
	assert.Equal(t, content, actual)

	source, _ = NewSource("foo", mustParseURL("env:/HELLO_WORLD"))

	actual, err = readEnv(source)
	assert.NoError(t, err)
	assert.Equal(t, content, actual)

	source, _ = NewSource("foo", mustParseURL("env:///HELLO_WORLD"))

	actual, err = readEnv(source)
	assert.NoError(t, err)
	assert.Equal(t, content, actual)

	source, _ = NewSource("foo", mustParseURL("env:HELLO_WORLD?foo=bar"))

	actual, err = readEnv(source)
	assert.NoError(t, err)
	assert.Equal(t, content, actual)

	source, _ = NewSource("foo", mustParseURL("env:///HELLO_WORLD?foo=bar"))

	actual, err = readEnv(source)
	assert.NoError(t, err)
	assert.Equal(t, content, actual)
}
