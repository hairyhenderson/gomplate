package xignore

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDirMatches_Simple(t *testing.T) {
	result, err := DirMatches("testdata/simple", &DirIgnoreOptions{
		IgnoreFile: ".xignore",
	})
	require.NoError(t, err)

	assert.Subset(t, result.MatchedFiles, []string{".xignore", "empty.log"})
	assert.Subset(t, result.UnmatchedFiles, []string{"rain.txt"})
	assert.Subset(t, result.MatchedDirs, []string{})
	assert.Subset(t, result.UnmatchedDirs, []string{})
}

func TestDirMatches_Nested(t *testing.T) {
	result, err := DirMatches("testdata/nested", &DirIgnoreOptions{
		IgnoreFile: ".xignore",
	})
	require.NoError(t, err)

	assert.Subset(t, result.MatchedFiles, []string{"inner/foo.md", "inner/2.lst"})
	assert.Subset(t, result.UnmatchedFiles, []string{".xignore", "1.txt", "inner/.xignore", "inner/inner2/moss.ini"})
	assert.Subset(t, result.MatchedDirs, []string{})
	assert.Subset(t, result.UnmatchedDirs, []string{"inner"})
}
