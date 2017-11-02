package xignore

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsMatchWildcard(t *testing.T) {
	match, _ := IsMatch("lost.txt", []string{"*"})
	if !match {
		t.Errorf("failed to get a wildcard match, got %v", match)
	}
}

// A simple pattern match should return true.
func TestIsMatch_Simple(t *testing.T) {
	match, _ := IsMatch("lost.txt", []string{"*.txt"})
	if !match {
		t.Errorf("failed to get a match, got %v", match)
	}
}

// An exclusion followed by an inclusion should return true.
func TestIsMatch_WithExclusion(t *testing.T) {
	match, _ := IsMatch("lost.txt", []string{"!lost.txt", "*.txt"})
	if !match {
		t.Errorf("failed to get true match on exclusion pattern, got %v", match)
	}
}

// A folder pattern followed by an exception should return false.
func TestIsMatch_WithFolderExclusion(t *testing.T) {
	match, _ := IsMatch("logs/README.md", []string{"logs", "!logs/README.md"})
	if match {
		t.Errorf("failed to get a false match on exclusion pattern, got %v", match)
	}
}

// A folder pattern followed by an exception should return false.
func TestIsMatch_WithSlashFolderExclusion(t *testing.T) {
	match, _ := IsMatch("logs/README.md", []string{"logs/", "!logs/README.md"})
	if match {
		t.Errorf("failed to get a false match on exclusion pattern, got %v", match)
	}
}

// A folder pattern followed by an exception should return false.
func TestIsMatch_WithWildcradFolderExclusion(t *testing.T) {
	match, _ := IsMatch("logs/README.md", []string{"logs/*", "!logs/README.md"})
	if match {
		t.Errorf("failed to get a false match on exclusion pattern, got %v", match)
	}
}

// A pattern followed by an exclusion should return false.
func TestIsMatch_WithAfterExclusion(t *testing.T) {
	match, _ := IsMatch("lost.txt", []string{"*.txt", "!lost.txt"})
	if match {
		t.Errorf("failed to get false match on exclusion pattern, got %v", match)
	}
}

// A filename evaluating to . should return false.
func TestIsMatch_ExclusionWholeDirectory(t *testing.T) {
	match, _ := IsMatch(".", []string{"*.txt"})
	if match {
		t.Errorf("failed to get false match on ., got %v", match)
	}
}

// Matches with no patterns
func TestIsMatch_WithNoPatterns(t *testing.T) {
	matches, err := IsMatch("/any/path/there", []string{})
	if err != nil {
		t.Fatal(err)
	}
	if matches {
		t.Fatalf("Should not have match anything")
	}
}

type matchesTestCase struct {
	pattern string
	text    string
	pass    bool
}

func TestIsMatch(t *testing.T) {
	tests := []matchesTestCase{
		{"**", "file", true},
		{"**", "file/", true},
		{"**/", "file", true}, // weird one
		{"**/", "file/", true},
		{"**", "/", true},
		{"**/", "/", true},
		{"**", "dir/file", true},
		{"**/", "dir/file", true},
		{"**", "dir/file/", true},
		{"**/", "dir/file/", true},
		{"**/**", "dir/file", true},
		{"**/**", "dir/file/", true},
		{"dir/**", "dir/file", true},
		{"dir/**", "dir/file/", true},
		{"dir/**", "dir/dir2/file", true},
		{"dir/**", "dir/dir2/file/", true},
		{"**/dir2/*", "dir/dir2/file", true},
		{"**/dir2/*", "dir/dir2/file/", true},
		{"**/dir2/**", "dir/dir2/dir3/file", true},
		{"**/dir2/**", "dir/dir2/dir3/file/", true},
		{"**file", "file", true},
		{"**file", "dir/file", true},
		{"**/file", "dir/file", true},
		{"**file", "dir/dir/file", true},
		{"**/file", "dir/dir/file", true},
		{"**/file*", "dir/dir/file", true},
		{"**/file*", "dir/dir/file.txt", true},
		{"**/file*txt", "dir/dir/file.txt", true},
		{"**/file*.txt", "dir/dir/file.txt", true},
		{"**/file*.txt*", "dir/dir/file.txt", true},
		{"**/**/*.txt", "dir/dir/file.txt", true},
		{"**/**/*.txt2", "dir/dir/file.txt", false},
		{"**/*.txt", "file.txt", true},
		{"**/**/*.txt", "file.txt", true},
		{"a**/*.txt", "a/file.txt", true},
		{"a**/*.txt", "a/dir/file.txt", true},
		{"a**/*.txt", "a/dir/dir/file.txt", true},
		{"a/*.txt", "a/dir/file.txt", false},
		{"a/*.txt", "a/file.txt", true},
		{"a/*.txt**", "a/file.txt", true},
		{"a[b-d]e", "ae", false},
		{"a[b-d]e", "ace", true},
		{"a[b-d]e", "aae", false},
		{"a[^b-d]e", "aze", true},
		{".*", ".foo", true},
		{".*", "foo", false},
		{"abc.def", "abcdef", false},
		{"abc.def", "abc.def", true},
		{"abc.def", "abcZdef", false},
		{"abc?def", "abcZdef", true},
		{"abc?def", "abcdef", false},
		{"a\\\\", "a\\", true},
		{"**/foo/bar", "foo/bar", true},
		{"**/foo/bar", "dir/foo/bar", true},
		{"**/foo/bar", "dir/dir2/foo/bar", true},
		{"abc/**", "abc", false},
		{"abc/**", "abc/def", true},
		{"abc/**", "abc/def/ghi", true},
		{"**/.foo", ".foo", true},
		{"**/.foo", "bar.foo", false},
		{"*.log", "dir/file.log", true},
	}

	if runtime.GOOS != "windows" {
		tests = append(tests, []matchesTestCase{
			{"a\\*b", "a*b", true},
			{"a\\", "a", false},
			{"a\\", "a\\", false},
		}...)
	}

	for _, test := range tests {
		desc := fmt.Sprintf("pattern=%q text=%q", test.pattern, test.text)
		pm := New([]string{test.pattern})
		res, _ := pm.Matches(test.text)
		assert.Equal(t, test.pass, res, desc)
	}
}

func TestCleanPatterns(t *testing.T) {
	patterns := []string{"docs", "config"}
	pm := New(patterns)
	cleaned := pm.Patterns()
	if len(cleaned) != 2 {
		t.Errorf("expected 2 element slice, got %v", len(cleaned))
	}
}

func TestCleanPatternsStripEmptyPatterns(t *testing.T) {
	patterns := []string{"docs", "config", ""}
	pm := New(patterns)
	cleaned := pm.Patterns()
	if len(cleaned) != 2 {
		t.Errorf("expected 2 element slice, got %v", len(cleaned))
	}
}

func TestCleanPatternsExceptionFlag(t *testing.T) {
	patterns := []string{"docs", "!docs/README.md"}
	pm := New(patterns)
	if !pm.HasExclusions() {
		t.Errorf("expected exceptions to be true, got %v", pm.HasExclusions())
	}
}

func TestCleanPatternsLeadingSpaceTrimmed(t *testing.T) {
	patterns := []string{"docs", "  !docs/README.md"}
	pm := New(patterns)
	if !pm.HasExclusions() {
		t.Errorf("expected exceptions to be true, got %v", pm.HasExclusions())
	}
}

func TestCleanPatternsTrailingSpaceTrimmed(t *testing.T) {
	patterns := []string{"docs", "!docs/README.md  "}
	pm := New(patterns)
	if !pm.HasExclusions() {
		t.Errorf("expected exceptions to be true, got %v", pm.HasExclusions())
	}
}

// These matchTests are stolen from go's filepath Match tests.
type matchTest struct {
	pattern, s string
	match      bool
	err        error
}

var matchTests = []matchTest{
	{"loot", "loot", true, nil},
	{"*", "fuck", true, nil},
	{"*d", "sword", true, nil},
	{"x*", "x", true, nil},
	{"a*", "avoid", true, nil},
	{"t*", "tan/cos", true, nil},
	{"m*/b", "min/b", true, nil},
	{"j*/k", "j/q/k", false, nil},
	{"a*b*c*d*e*/n", "axbxcxdxe/n", true, nil},
	{"a*b*c*d*e*/n", "axbxcxdxexxx/n", true, nil},
	{"a*b*c*d*e*/n", "axbxcxdxe/xxx/n", false, nil},
	{"a*b*c*d*e*/n", "axbxcxdxexxx/nnn", false, nil},
	{"a*b?c*x", "abxbbxdbxebxczzx", true, nil},
	{"a*b?c*x", "abxbbxdbxebxczzy", false, nil},
	{"ab[c]", "abc", true, nil},
	{"ab[b-d]", "abc", true, nil},
	{"ab[e-g]", "abc", false, nil},
	{"ab[^c]", "abc", false, nil},
	{"ab[^b-d]", "abc", false, nil},
	{"ab[^e-g]", "abc", true, nil},
	{"a\\*b", "a*b", true, nil},
	{"a\\*b", "ab", false, nil},
	{"a?b", "a☺b", true, nil},
	{"a[^a]b", "a☺b", true, nil},
	{"a???b", "a☺b", false, nil},
	{"a[^a][^a][^a]b", "a☺b", false, nil},
	{"[a-ζ]*", "α", true, nil},
	{"*[a-ζ]", "A", false, nil},
	{"a?b", "a/b", false, nil},
	{"a*b", "a/b", false, nil},
	{"[\\]a]", "]", true, nil},
	{"[\\-]", "-", true, nil},
	{"[x\\-]", "x", true, nil},
	{"[x\\-]", "-", true, nil},
	{"[x\\-]", "z", false, nil},
	{"[\\-x]", "x", true, nil},
	{"[\\-x]", "-", true, nil},
	{"[\\-x]", "a", false, nil},
	{"*x", "xxx", true, nil},
}

func errp(e error) string {
	if e == nil {
		return "<nil>"
	}
	return e.Error()
}

// TestMatch test's our version of filepath.Match, called regexpMatch.
func TestMatch(t *testing.T) {
	for _, tt := range matchTests {
		pattern := tt.pattern
		s := tt.s
		if runtime.GOOS == "windows" {
			if strings.Contains(pattern, "\\") {
				// no escape allowed on windows.
				continue
			}
			pattern = filepath.Clean(pattern)
			s = filepath.Clean(s)
		}
		ok, err := IsMatch(s, []string{pattern})
		if ok != tt.match || err != tt.err {
			t.Fatalf("Match(%#q, %#q) = %v, %q want %v, %q", pattern, s, ok, errp(err), tt.match, errp(tt.err))
		}
	}
}
