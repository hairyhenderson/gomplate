//go:build windows

package gomplate

import (
	"context"
	"testing"

	"github.com/hairyhenderson/gomplate/v3/internal/config"
	"github.com/spf13/afero"

	"github.com/stretchr/testify/assert"
)

func TestWalkDir(t *testing.T) {
	ctx := context.Background()
	origfs := fs
	defer func() { fs = origfs }()
	fs = afero.NewMemMapFs()

	cfg := &config.Config{}

	_, err := walkDir(ctx, cfg, `C:\indir`, simpleNamer(`C:\outdir`), nil, 0, false)
	assert.Error(t, err)

	_ = fs.MkdirAll(`C:\indir\one`, 0777)
	_ = fs.MkdirAll(`C:\indir\two`, 0777)
	afero.WriteFile(fs, `C:\indir\one\foo`, []byte("foo"), 0644)
	afero.WriteFile(fs, `C:\indir\one\bar`, []byte("bar"), 0644)
	afero.WriteFile(fs, `C:\indir\two\baz`, []byte("baz"), 0644)

	templates, err := walkDir(ctx, cfg, `C:\indir`, simpleNamer(`C:\outdir`), []string{`*\two`}, 0, false)

	assert.NoError(t, err)
	expected := []*tplate{
		{
			name:     `C:\indir\one\bar`,
			contents: "bar",
			mode:     0644,
		},
		{
			name:     `C:\indir\one\foo`,
			contents: "foo",
			mode:     0644,
		},
	}
	assert.Len(t, templates, 2)
	for i, tmpl := range templates {
		assert.Equal(t, expected[i].name, tmpl.name)
		assert.Equal(t, expected[i].contents, tmpl.contents)
		assert.Equal(t, expected[i].mode, tmpl.mode)
	}
}
