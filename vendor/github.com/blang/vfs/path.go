package vfs

import (
	"strings"
)

// SplitPath splits the given path in segments:
// 	"/" 				-> []string{""}
//	"./file" 			-> []string{".", "file"}
//	"file" 				-> []string{".", "file"}
//	"/usr/src/linux/" 	-> []string{"", "usr", "src", "linux"}
// The returned slice of path segments consists of one more more segments.
func SplitPath(path string, sep string) []string {
	path = strings.TrimSpace(path)
	path = strings.TrimSuffix(path, sep)
	if path == "" { // was "/"
		return []string{""}
	}
	if path == "." {
		return []string{"."}
	}

	if len(path) > 0 && !strings.HasPrefix(path, sep) && !strings.HasPrefix(path, "."+sep) {
		path = "./" + path
	}
	parts := strings.Split(path, sep)

	return parts
}
