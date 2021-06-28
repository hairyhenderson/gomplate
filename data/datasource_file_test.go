package data

import (
	"context"
	"testing"

	"github.com/hairyhenderson/gomplate/v3/internal/config"
	"github.com/hairyhenderson/gomplate/v3/internal/datasources"
	"github.com/spf13/afero"

	"github.com/stretchr/testify/assert"
)

func TestReadFile(t *testing.T) {
	content := []byte(`hello world`)
	fs := afero.NewMemMapFs()

	_ = fs.Mkdir("/tmp", 0777)
	f, _ := fs.Create("/tmp/foo")
	_, _ = f.Write(content)

	_ = fs.Mkdir("/tmp/partial", 0777)
	f, _ = fs.Create("/tmp/partial/foo.txt")
	_, _ = f.Write(content)
	_, _ = fs.Create("/tmp/partial/bar.txt")
	_, _ = fs.Create("/tmp/partial/baz.txt")

	source := config.DataSource{URL: mustParseURL("file:///tmp/foo")}

	ctx := config.WithFileSystem(context.Background(), fs)

	mt, actual, err := datasources.ReadDataSource(ctx, source)
	assert.NoError(t, err)
	assert.Equal(t, content, actual)
	assert.Equal(t, textMimetype, mt)

	source.URL = mustParseURL("file:///bogus")
	_, _, err = datasources.ReadDataSource(ctx, source)
	assert.Error(t, err)
	assert.Equal(t, textMimetype, mt)

	source.URL = mustParseURL("file:///tmp/bar")
	mt, actual, err = datasources.ReadDataSource(ctx, source, "foo")
	assert.NoError(t, err)
	assert.Equal(t, content, actual)
	assert.Equal(t, textMimetype, mt)

	source.URL = mustParseURL("file:///tmp/partial/")
	mt, actual, err = datasources.ReadDataSource(ctx, source, "foo.txt")
	assert.NoError(t, err)
	assert.Equal(t, content, actual)
	assert.Equal(t, textMimetype, mt)

	source.URL = mustParseURL("file:///tmp/partial/")
	mt, actual, err = datasources.ReadDataSource(ctx, source)
	assert.NoError(t, err)
	assert.Equal(t, []byte(`["bar.txt","baz.txt","foo.txt"]`), actual)
	assert.Equal(t, jsonArrayMimetype, mt)

	source.URL = mustParseURL("file:///tmp/partial/?type=application/json")
	mt, actual, err = datasources.ReadDataSource(ctx, source)
	assert.NoError(t, err)
	assert.Equal(t, []byte(`["bar.txt","baz.txt","foo.txt"]`), actual)
	assert.Equal(t, jsonMimetype, mt)

	source.URL = mustParseURL("file:///tmp/partial/?type=application/json")
	mt, actual, err = datasources.ReadDataSource(ctx, source, "foo.txt")
	assert.NoError(t, err)
	assert.Equal(t, content, actual)
	assert.Equal(t, jsonMimetype, mt)
}
