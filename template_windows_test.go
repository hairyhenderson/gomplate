//go:build windows

package gomplate

import (
	"context"
	"testing"

	"github.com/hairyhenderson/gomplate/v4/internal/config"
	"github.com/spf13/afero"

	"github.com/stretchr/testify/assert"
)

func TestWalkDir(t *testing.T) {
	ctx := context.Background()
	origfs := aferoFS
	defer func() { aferoFS = origfs }()
	aferoFS = afero.NewMemMapFs()

	cfg := &config.Config{}

	_, err := walkDir(ctx, cfg, `C:\indir`, simpleNamer(`C:\outdir`), nil, 0, false)
	assert.Error(t, err)

	_ = aferoFS.MkdirAll(`C:\indir\one`, 0777)
	_ = aferoFS.MkdirAll(`C:\indir\two`, 0777)
	afero.WriteFile(aferoFS, `C:\indir\one\foo`, []byte("foo"), 0644)
	afero.WriteFile(aferoFS, `C:\indir\one\bar`, []byte("bar"), 0644)
	afero.WriteFile(aferoFS, `C:\indir\two\baz`, []byte("baz"), 0644)

	templates, err := walkDir(ctx, cfg, `C:\indir`, simpleNamer(`C:\outdir`), []string{`*\two`}, 0, false)

	assert.NoError(t, err)
	expected := []Template{
		{
			Name: `C:\indir\one\bar`,
			Text: "bar",
		},
		{
			Name: `C:\indir\one\foo`,
			Text: "foo",
		},
	}
	assert.Len(t, templates, 2)
	for i, tmpl := range templates {
		assert.Equal(t, expected[i].Name, tmpl.Name)
		assert.Equal(t, expected[i].Text, tmpl.Text)
	}
}
