//go:build windows
// +build windows

package gomplate

import (
	"context"
	"io/fs"
	"os"
	"testing"

	"github.com/hack-pad/hackpadfs"
	"github.com/hack-pad/hackpadfs/mem"
	"github.com/hairyhenderson/gomplate/v4/internal/config"
	"github.com/hairyhenderson/gomplate/v4/internal/datafs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWalkDir_Windows(t *testing.T) {
	wd, _ := os.Getwd()
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})
	_ = os.Chdir("C:/")

	memfs, _ := mem.NewFS()
	fsys := datafs.WrapWdFS(memfs)

	ctx := datafs.ContextWithFSProvider(context.Background(), datafs.WrappedFSProvider(fsys, "file"))

	cfg := &config.Config{}

	_, err := walkDir(ctx, cfg, `C:\indir`, simpleNamer(`C:/outdir`), nil, nil, 0, false)
	require.Error(t, err)

	err = hackpadfs.MkdirAll(fsys, `C:\indir\one`, 0o777)
	require.NoError(t, err)
	err = hackpadfs.MkdirAll(fsys, `C:\indir\two`, 0o777)
	require.NoError(t, err)
	err = hackpadfs.WriteFullFile(fsys, `C:\indir\one\foo`, []byte("foo"), 0o644)
	require.NoError(t, err)
	err = hackpadfs.WriteFullFile(fsys, `C:\indir\one\bar`, []byte("bar"), 0o644)
	require.NoError(t, err)
	err = hackpadfs.WriteFullFile(fsys, `C:\indir\two\baz`, []byte("baz"), 0o644)
	require.NoError(t, err)

	fi, err := fs.Stat(fsys, `C:\indir\two\baz`)
	require.NoError(t, err)
	assert.Equal(t, "baz", fi.Name())

	templates, err := walkDir(ctx, cfg, `C:\indir`, simpleNamer(`C:/outdir`), []string{`*\two`}, []string{}, 0, false)
	require.NoError(t, err)

	expected := []Template{
		{
			Name: `C:/indir/one/bar`,
			Text: "bar",
		},
		{
			Name: `C:/indir/one/foo`,
			Text: "foo",
		},
	}
	require.Len(t, templates, 2)
	for i, tmpl := range templates {
		assert.Equal(t, expected[i].Name, tmpl.Name)
		assert.Equal(t, expected[i].Text, tmpl.Text)
	}
}
