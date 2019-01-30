package xignore

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/scanner"

	"github.com/spf13/afero"
)

// Pattern defines a single regexp used used to filter file paths.
type Pattern struct {
	value      string
	exclusion  bool
	regexpText string
	regexp     *regexp.Regexp
}

// NewPattern create new pattern
func NewPattern(strPattern string) *Pattern {
	if len(strPattern) == 0 {
		return &Pattern{value: ""} // empty
	}

	if strPattern[0] == '!' {
		if len(strPattern) == 1 {
			return &Pattern{value: ""} // empty
		}
		return &Pattern{value: strPattern[1:], exclusion: true}
	}

	return &Pattern{value: strPattern}
}

func (p *Pattern) String() string {
	strPattern := p.value
	if p.IsExclusion() {
		strPattern = "!" + strPattern
	}
	return p.value
}

// IsExclusion returns true if this pattern defines exclusion
func (p *Pattern) IsExclusion() bool {
	return p.exclusion
}

// IsEmpty returns true if this pattern is empty
func (p *Pattern) IsEmpty() bool {
	return p.value == ""
}

// IsRoot return true if this pattern is root
func (p *Pattern) IsRoot() bool {
	return len(p.value) > 0 && p.value[0] == os.PathSeparator
}

// Match match path
func (p *Pattern) Match(path string) bool {
	if p.regexp == nil {
		panic("regexp need compile")
	}

	if !strings.HasPrefix(path, string(os.PathSeparator)) {
		path = "/" + path
	}

	return p.regexp.MatchString(path) || p.regexp.MatchString(filepath.Base(path))
}

// Matches match paths
func (p *Pattern) Matches(files []string) []string {
	matchdFiles := []string{}
	for _, file := range files {
		if p.Match(file) {
			matchdFiles = append(matchdFiles, file)
		}
	}

	return matchdFiles
}

// Prepare preapre pattern
func (p *Pattern) Prepare() error {
	if p.regexp != nil {
		return nil
	}

	regStr := "^"
	pattern := p.value
	// Go through the pattern and convert it to a regexp.
	// We use a scanner so we can support utf-8 chars.
	var scan scanner.Scanner
	scan.Init(strings.NewReader(pattern))

	sl := string(os.PathSeparator)
	escSL := sl
	if sl == `\` {
		escSL += `\`
	}

	for scan.Peek() != scanner.EOF {
		ch := scan.Next()
		if scan.Pos().Offset == 1 && ch != '/' {
			// Optional root path
			regStr += (escSL + "?")
		}

		if ch == '*' {
			if scan.Peek() == '*' {
				// is some flavor of "**"
				scan.Next()

				// Treat **/ as ** so eat the "/"
				if string(scan.Peek()) == sl {
					scan.Next()
				}

				if scan.Peek() == scanner.EOF {
					// is "**EOF" - to align with .gitignore just accept all
					regStr += ".*"
				} else {
					// is "**"
					// Note that this allows for any # of /'s (even 0) because
					// the .* will eat everything, even /'s
					regStr += "(.*" + escSL + ")?"
				}
			} else {
				// is "*" so map it to anything but "/"
				regStr += "[^" + escSL + "]*"
			}
		} else if ch == '?' {
			// "?" is any char except "/"
			regStr += "[^" + escSL + "]"
		} else if ch == '.' || ch == '$' {
			// Escape some regexp special chars that have no meaning
			// in golang's filepath.Match
			regStr += `\` + string(ch)
		} else if ch == '\\' {
			// escape next char. Note that a trailing \ in the pattern
			// will be left alone (but need to escape it)
			if sl == `\` {
				// On windows map "\" to "\\", meaning an escaped backslash,
				// and then just continue because filepath.Match on
				// Windows doesn't allow escaping at all
				regStr += escSL
				continue
			}
			if scan.Peek() != scanner.EOF {
				regStr += `\` + string(scan.Next())
			} else {
				regStr += `\`
			}
		} else {
			regStr += string(ch)
		}
	}

	regStr += "$"

	re, err := regexp.Compile(regStr)
	if err != nil {
		return err
	}

	p.regexp = re
	p.regexpText = regStr
	return nil
}

func loadPatterns(vfs afero.Fs, ignorefile string) ([]*Pattern, error) {
	// read ignorefile
	ignoreFilePath := ignorefile
	if ignoreFilePath == "" {
		ignoreFilePath = DefaultIgnorefile
	}
	ignoreExists, err := afero.Exists(vfs, ignoreFilePath)
	if err != nil {
		return nil, err
	}

	// Load patterns from ignorefile
	patterns := []*Pattern{}
	if ignoreExists {
		f, err := vfs.Open(ignoreFilePath)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		ignoreFile := Ignorefile{}
		err = ignoreFile.FromReader(f)
		if err != nil {
			return nil, err
		}
		for _, sp := range ignoreFile.Patterns {
			pattern := NewPattern(sp)
			err := pattern.Prepare()
			if err != nil {
				return nil, err
			}
			patterns = append(patterns, pattern)
		}
	}

	return patterns, nil
}

func makePatterns(strPatterns []string) ([]*Pattern, error) {
	if strPatterns == nil || len(strPatterns) == 0 {
		return []*Pattern{}, nil
	}

	patterns := make([]*Pattern, len(strPatterns))
	for i, sp := range strPatterns {
		pattern := NewPattern(sp)
		if err := pattern.Prepare(); err != nil {
			return nil, err
		}
		patterns[i] = pattern
	}
	return patterns, nil
}
