package urlhelpers

import (
	"net/url"
	"path"
	"path/filepath"
)

// ParseSourceURL parses a datasource URL value, which may be '-' (for stdin://),
// or it may be a Windows path (with driver letter and back-slash separators) or
// UNC, or it may be relative. It also might just be a regular absolute URL...
// In all cases it returns a correct URL for the value. It may be a relative URL
// in which case the scheme should be assumed to be 'file'
func ParseSourceURL(value string) (*url.URL, error) {
	if value == "-" {
		value = "stdin://"
	}
	value = filepath.ToSlash(value)
	// handle absolute Windows paths
	volName := ""
	if volName = filepath.VolumeName(value); volName != "" {
		// handle UNCs
		if len(volName) > 2 {
			value = "file:" + value
		} else {
			value = "file:///" + value
		}
	}
	srcURL, err := url.Parse(value)
	if err != nil {
		return nil, err
	}

	if volName != "" && len(srcURL.Path) >= 3 {
		if srcURL.Path[0] == '/' && srcURL.Path[2] == ':' {
			srcURL.Path = srcURL.Path[1:]
		}
	}

	// if it's an absolute path with no scheme, assume it's a file
	if srcURL.Scheme == "" && path.IsAbs(srcURL.Path) {
		srcURL.Scheme = "file"
	}

	return srcURL, nil
}
