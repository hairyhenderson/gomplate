//go:build !windows
// +build !windows

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

	_, err := walkDir("/indir", simpleNamer("/outdir"), nil, 0, false)
	assert.Error(t, err)

	_ = osFS.MkdirAll("/indir/one", 0777)
	_ = osFS.MkdirAll("/indir/two", 0777)
	afero.WriteFile(osFS, "/indir/one/foo", []byte("foo"), 0644)
	afero.WriteFile(osFS, "/indir/one/bar", []byte("bar"), 0664)
	afero.WriteFile(osFS, "/indir/two/baz", []byte("baz"), 0644)

	templates, err := walkDir("/indir", simpleNamer("/outdir"), []string{"*/two"}, 0, false)

	assert.NoError(t, err)
	expected := []*tplate{
		{
			name:       "/indir/one/bar",
			targetPath: "/outdir/one/bar",
			mode:       0664,
		},
		{
			name:       "/indir/one/foo",
			targetPath: "/outdir/one/foo",
			mode:       0644,
		},
	}
	assert.EqualValues(t, expected, templates)
}
