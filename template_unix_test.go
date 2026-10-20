//go:build !windows

package gomplate

import (
	"context"
	"fmt"
	"io/fs"
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

func TestWalkDir_IncludeIgnored(t *testing.T) {
	memfs, _ := mem.NewFS()
	fsys := datafs.WrapWdFS(memfs)

	ctx := datafs.ContextWithFSProvider(context.Background(), datafs.WrappedFSProvider(fsys, "file"))

	require.NoError(t, hackpadfs.MkdirAll(fsys, "/indir/vendor", 0o777))
	require.NoError(t, hackpadfs.WriteFullFile(fsys, "/indir/.gomplateignore", []byte("vendor/\n"), 0o644))
	require.NoError(t, hackpadfs.WriteFullFile(fsys, "/indir/a.tmpl", []byte("a"), 0o644))
	require.NoError(t, hackpadfs.WriteFullFile(fsys, "/indir/vendor/b.tmpl", []byte("b"), 0o644))

	// with no negated patterns, files in the ignorefile are just excluded
	templates, err := walkDir(ctx, &Config{}, "/indir", simpleNamer("/outdir"), nil, nil, 0, false)
	require.NoError(t, err)
	assert.Contains(t, templateNames(templates), "/indir/a.tmpl")
	assert.NotContains(t, templateNames(templates), "/indir/vendor/b.tmpl")

	_, err = hackpadfs.Stat(fsys, "/outdir/vendor/b.tmpl")
	require.ErrorIs(t, err, fs.ErrNotExist)

	// a file that the ignorefile excludes but --include brings back is copied
	// without being rendered. This documents the current behaviour.
	templates, err = walkDir(ctx, &Config{}, "/indir", simpleNamer("/outdir"), []string{"*", "!*.tmpl"}, nil, 0, false)
	require.NoError(t, err)
	assert.Equal(t, []string{"/indir/a.tmpl"}, templateNames(templates))

	b, err := hackpadfs.ReadFile(fsys, "/outdir/vendor/b.tmpl")
	require.NoError(t, err)
	assert.Equal(t, "b", string(b))
}

func templateNames(templates []Template) []string {
	names := make([]string, len(templates))
	for i, t := range templates {
		names[i] = t.Name
	}

	return names
}

var walkDirSink []Template

func BenchmarkWalkDir(b *testing.B) {
	memfs, _ := mem.NewFS()
	fsys := datafs.WrapWdFS(memfs)

	ctx := datafs.ContextWithFSProvider(context.Background(), datafs.WrappedFSProvider(fsys, "file"))

	// 1000 files in 10 directories, with an ignorefile
	require.NoError(b, hackpadfs.MkdirAll(fsys, "/indir", 0o777))
	require.NoError(b, hackpadfs.WriteFullFile(fsys, "/indir/.gomplateignore", []byte("*.bak\n"), 0o644))

	for d := range 10 {
		dir := fmt.Sprintf("/indir/d%d", d)
		require.NoError(b, hackpadfs.MkdirAll(fsys, dir, 0o777))

		for f := range 100 {
			name := fmt.Sprintf("%s/f%d.txt", dir, f)
			require.NoError(b, hackpadfs.WriteFullFile(fsys, name, []byte("hello"), 0o644))
		}
	}

	cfg := &Config{}

	for b.Loop() {
		var err error
		walkDirSink, err = walkDir(ctx, cfg, "/indir", simpleNamer("/outdir"), nil, nil, 0, false)
		if err != nil {
			b.Fatal(err)
		}
	}
}
