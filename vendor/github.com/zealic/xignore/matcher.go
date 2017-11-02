package xignore

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/scanner"
)

// IgnoreMatcher allows checking paths agaist a list of patterns
type IgnoreMatcher struct {
	patterns     []*Pattern
	hasExclusion bool
}

// New creates a new matcher object for specific patterns that can
// be used later to match against patterns against paths
func New(patterns []string) *IgnoreMatcher {
	im := &IgnoreMatcher{
		patterns: make([]*Pattern, 0, len(patterns)),
	}
	for _, sp := range patterns {
		sp = strings.TrimSpace(sp)
		if sp == "" {
			continue
		}
		sp = filepath.Clean(sp)
		pattern := &Pattern{}
		if sp[0] == '!' {
			if len(sp) == 1 {
				continue
			}
			pattern.exclusion = true
			sp = sp[1:]
			im.hasExclusion = true
		}
		if _, err := filepath.Match(sp, "."); err != nil {
			continue
		}
		pattern.value = sp
		pattern.dirs = strings.Split(sp, string(os.PathSeparator))
		im.patterns = append(im.patterns, pattern)
	}
	return im
}

// Matches matches path against all the patterns. Matches is not safe to be
// called concurrently
func (im *IgnoreMatcher) Matches(file string) (bool, error) {
	file = filepath.FromSlash(file)
	parentPath := filepath.Dir(file)
	parentPathDirs := strings.Split(parentPath, string(os.PathSeparator))

	matched := false
	for _, pattern := range im.patterns {
		match, err := pattern.Match(file)
		if err != nil {
			return false, err
		}

		if !match && parentPath != "." {
			// Check to see if the pattern matches one of our parent dirs.
			if len(pattern.dirs) <= len(parentPathDirs) {
				match, _ = pattern.Match(
					strings.Join(parentPathDirs[:len(pattern.dirs)],
						string(os.PathSeparator)))
			}
		}

		if match {
			matched = !pattern.exclusion
		}
	}

	return matched, nil
}

// HasExclusions returns true if any pattern define exclusion
func (im *IgnoreMatcher) HasExclusions() bool {
	return im.hasExclusion
}

// Patterns returns array of active patterns
func (im *IgnoreMatcher) Patterns() []*Pattern {
	return im.patterns
}

// Pattern defines a single regexp used used to filter file paths.
type Pattern struct {
	value     string
	dirs      []string
	regexp    *regexp.Regexp
	exclusion bool
}

func (p *Pattern) String() string {
	return p.value
}

// Exclusion returns true if this pattern defines exclusion
func (p *Pattern) Exclusion() bool {
	return p.exclusion
}

// Match match path
func (p *Pattern) Match(path string) (bool, error) {
	if p.regexp == nil {
		if err := p.compile(); err != nil {
			return false, filepath.ErrBadPattern
		}
	}

	b := p.regexp.MatchString(path) || p.regexp.MatchString(filepath.Base(path))

	return b, nil
}

func (p *Pattern) compile() error {
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
	return nil
}
