//go:build !windows

package gomplate

import (
	"context"
	"testing"

	"github.com/hairyhenderson/gomplate/v4/internal/config"
	"github.com/spf13/afero"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWalkDir(t *testing.T) {
	ctx := context.Background()
	origfs := aferoFS
	defer func() { aferoFS = origfs }()
	aferoFS = afero.NewMemMapFs()

	cfg := &config.Config{}

	_, err := walkDir(ctx, cfg, "/indir", simpleNamer("/outdir"), nil, 0, false)
	assert.Error(t, err)

	_ = aferoFS.MkdirAll("/indir/one", 0777)
	_ = aferoFS.MkdirAll("/indir/two", 0777)
	afero.WriteFile(aferoFS, "/indir/one/foo", []byte("foo"), 0644)
	afero.WriteFile(aferoFS, "/indir/one/bar", []byte("bar"), 0664)
	afero.WriteFile(aferoFS, "/indir/two/baz", []byte("baz"), 0644)

	templates, err := walkDir(ctx, cfg, "/indir", simpleNamer("/outdir"), []string{"*/two"}, 0, false)

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
