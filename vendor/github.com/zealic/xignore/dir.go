package xignore

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

// DirIgnoreOptions dir ignore options
type DirIgnoreOptions struct {
	// Base directory, default to use 'dir' parameter
	BaseDir string
	// Ignorefile name, similar '.gitignore', '.dockerignore', 'chefignore'
	IgnoreFile string
	// Before user patterns.
	BeforePatterns []string
	// After user patterns.
	AfterPatterns []string
	// // Global ignore filename, similar '.gitignore_global'
	// GlobalIgnoreFile string
	// // No inherit parent directory ignorefile
	// Isolate bool
}

// DirMatchesResult directory matches result
type DirMatchesResult struct {
	BaseDir        string
	MatchedFiles   []string
	UnmatchedFiles []string
	MatchedDirs    []string
	UnmatchedDirs  []string
	//ErrorFiles     []string
}

// DirMatches returns matched files from dir files.
func DirMatches(dir string, options *DirIgnoreOptions) (*DirMatchesResult, error) {
	var err error
	basedir := options.BaseDir
	if basedir == "" {
		basedir = dir
	}
	if !filepath.IsAbs(basedir) {
		basedir, err = filepath.Abs(basedir)
		if err != nil {
			return nil, err
		}
	}
	if !filepath.IsAbs(dir) {
		dir, err = filepath.Abs(dir)
		if err != nil {
			return nil, err
		}
	}

	// assert tha dir exists
	_, err = os.Stat(dir)
	if err != nil {
		return nil, err
	}

	// read directory
	entries, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	ignoreFilePath := filepath.Join(dir, options.IgnoreFile)
	patterns, err := flatPatterns(ignoreFilePath, options)
	if err != nil {
		return nil, err
	}
	matcher := New(patterns)

	// process or dive in again
	mfiles := []string{}
	ufiles := []string{}
	mdirs := []string{}
	udirs := []string{}
	for _, entry := range entries {
		subpath := filepath.Join(dir, entry.Name())
		relpath, err := filepath.Rel(basedir, subpath)

		if err != nil {
			return nil, err
		}

		match, err := matcher.Matches(relpath)
		if err != nil {
			return nil, err
		}

		if entry.IsDir() {
			subResult, err := DirMatches(filepath.Join(dir, entry.Name()), &DirIgnoreOptions{
				BaseDir:        basedir,
				IgnoreFile:     options.IgnoreFile,
				BeforePatterns: patterns,
			})
			if err != nil {
				return nil, err
			}

			if match {
				mdirs = append(mdirs, relpath)
			} else {
				udirs = append(udirs, relpath)
			}
			mfiles = append(mfiles, subResult.MatchedFiles...)
			ufiles = append(ufiles, subResult.UnmatchedFiles...)
			mdirs = append(mdirs, subResult.MatchedDirs...)
			udirs = append(udirs, subResult.UnmatchedDirs...)
		} else {
			if match {
				mfiles = append(mfiles, relpath)
			} else {
				ufiles = append(ufiles, relpath)
			}
		}
	}
	return &DirMatchesResult{
		BaseDir:        basedir,
		MatchedFiles:   mfiles,
		UnmatchedFiles: ufiles,
		MatchedDirs:    mdirs,
		UnmatchedDirs:  udirs,
	}, nil
}

func flatPatterns(ignoreFilePath string, options *DirIgnoreOptions) ([]string, error) {
	patterns := []string{}
	if options.BeforePatterns != nil {
		patterns = append(patterns, options.BeforePatterns...)
	}

	if _, err := os.Stat(ignoreFilePath); !os.IsNotExist(err) {

		f, err := os.Open(ignoreFilePath)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		ignoreFile := Ignorefile{}
		err = ignoreFile.FromReader(f)
		if err != nil {
			return nil, err
		}
		patterns = append(patterns, ignoreFile.Patterns...)
	}

	if options.AfterPatterns != nil {
		patterns = append(patterns, options.AfterPatterns...)
	}

	return patterns, nil
}
