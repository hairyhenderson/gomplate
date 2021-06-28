package data

import (
	"context"
	"net/url"
	"os"
	"testing"

	"github.com/hairyhenderson/gomplate/v3/internal/config"
	"github.com/hairyhenderson/gomplate/v3/internal/datasources"
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

	source := config.DataSource{URL: mustParseURL("env:HELLO_WORLD")}

	ctx := context.Background()

	_, actual, err := datasources.ReadDataSource(ctx, source)
	assert.NoError(t, err)
	assert.Equal(t, content, actual)

	source = config.DataSource{URL: mustParseURL("env:/HELLO_WORLD")}

	_, actual, err = datasources.ReadDataSource(ctx, source)
	assert.NoError(t, err)
	assert.Equal(t, content, actual)

	source = config.DataSource{URL: mustParseURL("env:///HELLO_WORLD")}

	_, actual, err = datasources.ReadDataSource(ctx, source)
	assert.NoError(t, err)
	assert.Equal(t, content, actual)

	source = config.DataSource{URL: mustParseURL("env:HELLO_WORLD?foo=bar")}

	_, actual, err = datasources.ReadDataSource(ctx, source)
	assert.NoError(t, err)
	assert.Equal(t, content, actual)

	source = config.DataSource{URL: mustParseURL("env:///HELLO_WORLD?foo=bar")}

	_, actual, err = datasources.ReadDataSource(ctx, source)
	assert.NoError(t, err)
	assert.Equal(t, content, actual)
}
