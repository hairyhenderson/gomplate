package data

import (
	"net/url"
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func mustParseURL(in string) *url.URL {
	u, _ := url.Parse(in)
	p := regexp.MustCompile("^/[a-zA-Z]:.*$")
	if p.MatchString(u.Path) {
		u.Path = trimLeftChar(u.Path)
	}
	return u
}

func TestReadEnv(t *testing.T) {
	content := []byte(`hello world`)
	os.Setenv("HELLO_WORLD", "hello world")
	defer os.Unsetenv("HELLO_WORLD")
	os.Setenv("HELLO_UNIVERSE", "hello universe")
	defer os.Unsetenv("HELLO_UNIVERSE")

	source := &Source{Alias: "foo", URL: mustParseURL("env:HELLO_WORLD")}

	actual, err := readEnv(source)
	assert.NoError(t, err)
	assert.Equal(t, content, actual)

	source = &Source{Alias: "foo", URL: mustParseURL("env:/HELLO_WORLD")}

	actual, err = readEnv(source)
	assert.NoError(t, err)
	assert.Equal(t, content, actual)

	source = &Source{Alias: "foo", URL: mustParseURL("env:///HELLO_WORLD")}

	actual, err = readEnv(source)
	assert.NoError(t, err)
	assert.Equal(t, content, actual)

	source = &Source{Alias: "foo", URL: mustParseURL("env:HELLO_WORLD?foo=bar")}

	actual, err = readEnv(source)
	assert.NoError(t, err)
	assert.Equal(t, content, actual)

	source = &Source{Alias: "foo", URL: mustParseURL("env:///HELLO_WORLD?foo=bar")}

	actual, err = readEnv(source)
	assert.NoError(t, err)
	assert.Equal(t, content, actual)
}
