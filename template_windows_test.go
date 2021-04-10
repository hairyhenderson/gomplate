//go:build windows
// +build windows

package gomplate

import (
	"testing"

	"github.com/spf13/afero"

	"github.com/stretchr/testify/assert"
)

func TestWalkDir(t *testing.T) {
	origfs := osFS
	defer func() { osFS = origfs }()
	osFS = afero.NewMemMapFs()

	_, err := walkDir(`C:\indir`, simpleNamer(`C:\outdir`), nil, 0, false)
	assert.Error(t, err)

	_ = osFS.MkdirAll(`C:\indir\one`, 0777)
	_ = osFS.MkdirAll(`C:\indir\two`, 0777)
	afero.WriteFile(osFS, `C:\indir\one\foo`, []byte("foo"), 0644)
	afero.WriteFile(osFS, `C:\indir\one\bar`, []byte("bar"), 0644)
	afero.WriteFile(osFS, `C:\indir\two\baz`, []byte("baz"), 0644)

	templates, err := walkDir(`C:\indir`, simpleNamer(`C:\outdir`), []string{`*\two`}, 0, false)

	assert.NoError(t, err)
	expected := []*tplate{
		{
			name:       `C:\indir\one\bar`,
			targetPath: `C:\outdir\one\bar`,
			mode:       0644,
		},
		{
			name:       `C:\indir\one\foo`,
			targetPath: `C:\outdir\one\foo`,
			mode:       0644,
		},
	}
	assert.EqualValues(t, expected, templates)
}
