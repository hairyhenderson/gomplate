package datasource

import (
	"context"
	"testing"

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

	ctx := context.Background()
	ff := &File{fs}

	actual, err := ff.Read(ctx, mustParseURL("file:///tmp/foo"))
	assert.NoError(t, err)
	assert.Equal(t, content, actual.Bytes)
	// assert.Equal(t, "", actual.MediaType)

	_, err = ff.Read(ctx, mustParseURL("file:///bogus"))
	assert.Error(t, err)

	actual, err = ff.Read(ctx, mustParseURL("file:///tmp/partial"), "foo.txt")
	assert.NoError(t, err)
	assert.Equal(t, content, actual.Bytes)

	actual, err = ff.Read(ctx, mustParseURL("file:///tmp/partial/"))
	assert.NoError(t, err)
	assert.Equal(t, []byte(`["bar.txt","baz.txt","foo.txt"]`), actual.Bytes)

	actual, err = ff.Read(ctx, mustParseURL("file:///tmp/partial/?type=application/json"))
	assert.NoError(t, err)
	assert.Equal(t, []byte(`["bar.txt","baz.txt","foo.txt"]`), actual.Bytes)
	assert.Equal(t, jsonMimetype, must(actual.MediaType()))

	actual, err = ff.Read(ctx, mustParseURL("file:///tmp/partial/?type=application/json"), "foo.txt")
	assert.NoError(t, err)
	assert.Equal(t, content, actual.Bytes)
	assert.Equal(t, jsonMimetype, must(actual.MediaType()))
}
