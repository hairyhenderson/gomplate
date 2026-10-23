package funcs

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateFilePathFuncs(t *testing.T) {
	t.Parallel()

	for i := range 10 {
		// Run this a bunch to catch race conditions
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			fmap := CreateFilePathFuncs(ctx)
			actual := fmap["filepath"].(func() any)

			assert.Equal(t, ctx, actual().(*FilePathFuncs).ctx)
		})
	}
}

func TestFilePathGlob(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("test"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "file2.log"), []byte("test"), 0o600))

	ff := &FilePathFuncs{ctx: context.Background()}

	matches, err := ff.Glob(filepath.Join(tmpDir, "*.txt"))
	require.NoError(t, err)
	assert.Equal(t, []string{filepath.Join(tmpDir, "file1.txt")}, matches)

	// Malformed pattern test
	_, err = ff.Glob("[]")
	assert.Error(t, err)
}
