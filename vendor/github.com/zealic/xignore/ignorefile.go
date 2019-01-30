package xignore

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

// DefaultIgnorefile default ignorefile name ".xignore"
const DefaultIgnorefile = ".xignore"

// Ignorefile ignore file
type Ignorefile struct {
	Patterns []string
}

// FromReader reads patterns from reader.
// This will trim whitespace from each line as well
// as use GO's "clean" func to get the shortest/cleanest path for each.
func (f *Ignorefile) FromReader(reader io.Reader) error {
	if reader == nil {
		return nil
	}

	scanner := bufio.NewScanner(reader)
	var excludes []string
	currentLine := 0

	utf8bom := []byte{0xEF, 0xBB, 0xBF}
	for scanner.Scan() {
		scannedBytes := scanner.Bytes()
		// TRIM UTF8 BOM
		if currentLine == 0 {
			scannedBytes = bytes.TrimPrefix(scannedBytes, utf8bom)
		}
		pattern := string(scannedBytes)
		currentLine++
		// Lines starting with # (comments) are ignored before processing
		if strings.HasPrefix(pattern, "#") {
			continue
		}
		pattern = strings.TrimRight(pattern, " ")
		if pattern == "" {
			continue
		}
		// normalize absolute paths to paths relative to the context
		// (taking care of '!' prefix)
		invert := pattern[0] == '!'
		if invert {
			pattern = strings.TrimRight(pattern[1:], " ")
		}
		if len(pattern) > 0 {
			pattern = filepath.Clean(pattern)
			pattern = filepath.ToSlash(pattern)
		}
		if invert {
			pattern = "!" + pattern
		}

		excludes = append(excludes, pattern)
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("Error reading ignorefile: %v", err)
	}

	f.Patterns = excludes
	return nil
}
