package xignore

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/afero"
)

type stateMap map[string]bool

func collectFiles(fs afero.Fs) (files []string, err error) {
	files = []string{}

	afero.Walk(fs, "", func(path string, info os.FileInfo, werr error) error {
		if werr != nil {
			err = werr
			return nil
		}
		files = append(files, path)
		return nil
	})
	return
}

func (state stateMap) merge(source stateMap) {
	for k, val := range source {
		state[k] = val
	}
}

func (state stateMap) mergeFiles(files []string, value bool) {
	for _, f := range files {
		state[f] = value
	}
}

func (state stateMap) applyPatterns(vfs afero.Fs, files []string, patterns []*Pattern) error {
	filesMap := stateMap{}
	dirPatterns := []*Pattern{}
	for _, pattern := range patterns {
		if pattern.IsEmpty() {
			continue
		}
		currFiles := pattern.Matches(files)
		if pattern.IsExclusion() {
			for _, f := range currFiles {
				filesMap[f] = false
			}
		} else {
			for _, f := range currFiles {
				filesMap[f] = true
			}
		}

		// generate dir based patterns
		for _, f := range currFiles {
			ok, err := afero.IsDir(vfs, f)
			if err != nil {
				return err
			}
			if ok {
				strPattern := f + "/**"
				if pattern.IsExclusion() {
					strPattern = "!" + strPattern
				}
				dirPattern := NewPattern(strPattern)
				dirPatterns = append(dirPatterns, dirPattern)
				err := dirPattern.Prepare()
				if err != nil {
					return err
				}
			}
		}
	}

	// handle dirs batch matches
	dirFileMap := stateMap{}
	for _, pattern := range dirPatterns {
		if pattern.IsEmpty() {
			continue
		}
		currFiles := pattern.Matches(files)
		if pattern.IsExclusion() {
			for _, f := range currFiles {
				dirFileMap[f] = false
			}
		} else {
			for _, f := range currFiles {
				dirFileMap[f] = true
			}
		}
	}

	state.merge(dirFileMap)
	state.merge(filesMap)
	return nil
}

func (state stateMap) applyIgnorefile(vfs afero.Fs, ignorefile string, nested bool) error {
	// Apply nested ignorefile
	ignorefiles := []string{}

	if nested {
		for file := range state {
			// all subdir ignorefiles
			if strings.HasSuffix(file, ignorefile) {
				ignorefiles = append(ignorefiles, file)
			}
		}
		// Sort by dir deep level
		sort.Slice(ignorefiles, func(i, j int) bool {
			ilen := len(strings.Split(ignorefiles[i], string(os.PathSeparator)))
			jlen := len(strings.Split(ignorefiles[j], string(os.PathSeparator)))
			return ilen < jlen
		})
	} else {
		ignorefiles = []string{ignorefile}
	}

	for _, ifile := range ignorefiles {
		currBasedir := filepath.Dir(ifile)
		currFs := vfs
		if currBasedir != "." {
			currFs = afero.NewBasePathFs(vfs, currBasedir)
		}
		patterns, err := loadPatterns(currFs, ignorefile)
		if err != nil {
			return err
		}

		currMap := stateMap{}
		currFiles, err := collectFiles(currFs)
		if err != nil {
			return err
		}
		err = currMap.applyPatterns(currFs, currFiles, patterns)
		if err != nil {
			return err
		}

		for nfile, matched := range currMap {
			parentFile := filepath.Join(currBasedir, nfile)
			state[parentFile] = matched
		}
	}

	return nil
}
