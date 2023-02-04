package data

import (
	"context"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/hairyhenderson/gomplate/v4/internal/datafs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadFile(t *testing.T) {
	ctx := context.Background()

	content := []byte(`hello world`)

	fsys := datafs.WrapWdFS(fstest.MapFS{
		"tmp":                 {Mode: fs.ModeDir | 0o777},
		"tmp/foo":             {Data: content},
		"tmp/partial":         {Mode: fs.ModeDir | 0o777},
		"tmp/partial/foo.txt": {Data: content},
		"tmp/partial/bar.txt": {},
		"tmp/partial/baz.txt": {},
	})

	source := &Source{Alias: "foo", URL: mustParseURL("file:///tmp/foo")}
	source.fs = fsys

	actual, err := readFile(ctx, source)
	require.NoError(t, err)
	assert.Equal(t, content, actual)

	source = &Source{Alias: "bogus", URL: mustParseURL("file:///bogus")}
	source.fs = fsys
	_, err = readFile(ctx, source)
	assert.Error(t, err)

	source = &Source{Alias: "partial", URL: mustParseURL("file:///tmp/partial")}
	source.fs = fsys
	actual, err = readFile(ctx, source, "foo.txt")
	require.NoError(t, err)
	assert.Equal(t, content, actual)

	source = &Source{Alias: "dir", URL: mustParseURL("file:///tmp/partial/")}
	source.fs = fsys
	actual, err = readFile(ctx, source)
	require.NoError(t, err)
	assert.Equal(t, []byte(`["bar.txt","baz.txt","foo.txt"]`), actual)

	source = &Source{Alias: "dir", URL: mustParseURL("file:///tmp/partial/?type=application/json")}
	source.fs = fsys
	actual, err = readFile(ctx, source)
	require.NoError(t, err)
	assert.Equal(t, []byte(`["bar.txt","baz.txt","foo.txt"]`), actual)
	mime, err := source.mimeType("")
	require.NoError(t, err)
	assert.Equal(t, "application/json", mime)

	source = &Source{Alias: "dir", URL: mustParseURL("file:///tmp/partial/?type=application/json")}
	source.fs = fsys
	actual, err = readFile(ctx, source, "foo.txt")
	require.NoError(t, err)
	assert.Equal(t, content, actual)
	mime, err = source.mimeType("")
	require.NoError(t, err)
	assert.Equal(t, "application/json", mime)
}
