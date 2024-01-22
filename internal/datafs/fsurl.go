package datafs

import (
	"net/url"
	"strings"
)

// SplitFSMuxURL splits a URL into a filesystem URL and a relative file path
func SplitFSMuxURL(in *url.URL) (*url.URL, string) {
	u := *in

	// git URLs are special - they have double-slashes that separate a repo
	// from a path in the repo. A missing double-slash means the path is the
	// root.
	switch u.Scheme {
	case "git", "git+file", "git+http", "git+https", "git+ssh":
		repo, base, _ := strings.Cut(u.Path, "//")
		u.Path = repo
		if base == "" {
			base = "."
		}

		return &u, base
	}

	// trim leading and trailing slashes - they are not part of a valid path
	// according to [io/fs.ValidPath]
	base := strings.Trim(u.Path, "/")

	if base == "" && u.Opaque != "" {
		base = u.Opaque
		u.Opaque = ""
	}

	if base == "" {
		base = "."
	}

	u.Path = "/"

	return &u, base
}
