//go:build !windows

package gomplate

import (
	"context"
	"testing"

	"github.com/hack-pad/hackpadfs"
	"github.com/hack-pad/hackpadfs/mem"
	"github.com/hairyhenderson/gomplate/v5/internal/datafs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWalkDir_UNIX(t *testing.T) {
	memfs, _ := mem.NewFS()
	fsys := datafs.WrapWdFS(memfs)

	ctx := datafs.ContextWithFSProvider(context.Background(), datafs.WrappedFSProvider(fsys, "file"))

	cfg := &Config{}

	_, err := walkDir(ctx, cfg, "/indir", simpleNamer("/outdir"), nil, nil, 0, false)
	require.Error(t, err)

	err = hackpadfs.MkdirAll(fsys, "/indir/one", 0o777)
	require.NoError(t, err)
	err = hackpadfs.MkdirAll(fsys, "/indir/two", 0o777)
	require.NoError(t, err)
	err = hackpadfs.WriteFullFile(fsys, "/indir/one/foo", []byte("foo"), 0o644)
	require.NoError(t, err)
	err = hackpadfs.WriteFullFile(fsys, "/indir/one/bar", []byte("bar"), 0o644)
	require.NoError(t, err)
	err = hackpadfs.WriteFullFile(fsys, "/indir/two/baz", []byte("baz"), 0o644)
	require.NoError(t, err)

	templates, err := walkDir(ctx, cfg, "/indir", simpleNamer("/outdir"), []string{"*/two"}, []string{}, 0, false)
	require.NoError(t, err)

	expected := []Template{
		{
			Name: "/indir/one/bar",
			Text: "bar",
		},
		{
			Name: "/indir/one/foo",
			Text: "foo",
		},
	}
	assert.Len(t, templates, 2)
	for i, tmpl := range templates {
		assert.Equal(t, expected[i].Name, tmpl.Name)
		assert.Equal(t, expected[i].Text, tmpl.Text)
	}
}
