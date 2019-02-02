package xignore

// MatchesOptions matches options
type MatchesOptions struct {
	// Ignorefile name, similar '.gitignore', '.dockerignore', 'chefignore'
	Ignorefile string
	// Allow nested ignorefile
	Nested bool
	// apply patterns before all ignorefile
	BeforePatterns []string
	// apply patterns after all ignorefile
	AfterPatterns []string
}

// MatchesResult matches result
type MatchesResult struct {
	BaseDir string
	// ignorefile rules matched files
	MatchedFiles []string
	// ignorefile rules unmatched files
	UnmatchedFiles []string
	// ignorefile rules matched dirs
	MatchedDirs []string
	// ignorefile rules unmatched dirs
	UnmatchedDirs []string
}

// DirMatches returns match result from basedir.
func DirMatches(basedir string, options *MatchesOptions) (*MatchesResult, error) {
	return NewSystemMatcher().Matches(basedir, options)
}
