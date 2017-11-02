package xignore

import (
	"path/filepath"
)

// IsMatch returns true if file matches any of the patterns
// and isn't excluded by any of the subsequent patterns.
func IsMatch(file string, patterns []string) (bool, error) {
	im := New(patterns)
	file = filepath.Clean(file)

	if file == "." {
		// Don't let them exclude everything, kind of silly.
		return false, nil
	}

	return im.Matches(file)
}
