package datafs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/hairyhenderson/go-fsimpl"
	"github.com/hairyhenderson/gomplate/v4/internal/config"
	"github.com/hairyhenderson/gomplate/v4/internal/iohelpers"
)

const osWindows = "windows"

// typeOverrideParam gets the query parameter used to override the content type
// used to parse a given datasource - use GOMPLATE_TYPE_PARAM to use a different
// parameter name.
func typeOverrideParam() string {
	if v := os.Getenv("GOMPLATE_TYPE_PARAM"); v != "" {
		return v
	}

	return "type"
}

// DataSourceReader reads content from a datasource
type DataSourceReader interface {
	// ReadSource reads the content of a datasource, given an alias and optional
	// arguments. If the datasource is not found, the alias is interpreted as a
	// URL. If the alias is not a valid URL, an error is returned.
	//
	// Returned content is cached, so subsequent calls with the same alias and
	// arguments will return the same content.
	ReadSource(ctx context.Context, alias string, args ...string) (string, []byte, error)

	// contains registry
	Registry
}

type dsReader struct {
	cache map[string]*content

	Registry
}

// content type mainly for caching
type content struct {
	contentType string
	b           []byte
}

func NewSourceReader(reg Registry) DataSourceReader {
	return &dsReader{Registry: reg}
}

func (d *dsReader) ReadSource(ctx context.Context, alias string, args ...string) (string, []byte, error) {
	source, ok := d.Lookup(alias)
	if !ok {
		srcURL, err := url.Parse(alias)
		if err != nil || !srcURL.IsAbs() {
			return "", nil, fmt.Errorf("undefined datasource '%s': %w", alias, err)
		}

		d.Register(alias, config.DataSource{URL: srcURL})

		// repeat the lookup now that it's registered - we shouldn't just use
		// it directly because registration may include extra headers
		source, _ = d.Lookup(alias)
	}

	if d.cache == nil {
		d.cache = make(map[string]*content)
	}
	cacheKey := alias
	for _, v := range args {
		cacheKey += v
	}
	cached, ok := d.cache[cacheKey]
	if ok {
		return cached.contentType, cached.b, nil
	}

	arg := ""
	if len(args) > 0 {
		arg = args[0]
	}
	u, err := resolveURL(*source.URL, arg)
	if err != nil {
		return "", nil, err
	}

	fc, err := d.readFileContent(ctx, u, source.Header)
	if err != nil {
		return "", nil, fmt.Errorf("couldn't read datasource '%s' (%s): %w", alias, u, err)
	}
	d.cache[cacheKey] = fc

	return fc.contentType, fc.b, nil
}

func removeQueryParam(u *url.URL, key string) *url.URL {
	q := u.Query()
	q.Del(key)
	u.RawQuery = q.Encode()
	return u
}

func (d *dsReader) readFileContent(ctx context.Context, u *url.URL, hdr http.Header) (*content, error) {
	// possible type hint in the type query param. Contrary to spec, we allow
	// unescaped '+' characters to make it simpler to provide types like
	// "application/array+json"
	overrideType := typeOverrideParam()
	mimeType := u.Query().Get(overrideType)
	mimeType = strings.ReplaceAll(mimeType, " ", "+")

	// now that we have the hint, remove it from the URL - we can't have it
	// leaking into the filesystem layer
	u = removeQueryParam(u, overrideType)

	u, fname := SplitFSMuxURL(u)

	fsys, err := FSysForPath(ctx, u.String())
	if err != nil {
		return nil, fmt.Errorf("fsys for path %v: %w", u, err)
	}

	// need to support absolute paths on local filesystem too
	// TODO: this is a hack, probably fix this?
	if u.Scheme == "file" && runtime.GOOS != osWindows {
		fname = u.Path + fname
	}

	fsys = fsimpl.WithContextFS(ctx, fsys)
	fsys = fsimpl.WithHeaderFS(hdr, fsys)
	fsys = WithDataSourceRegistryFS(d.Registry, fsys)

	f, err := fsys.Open(fname)
	if err != nil {
		return nil, fmt.Errorf("open (url: %q, name: %q): %w", u, fname, err)
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat (url: %q, name: %q): %w", u, fname, err)
	}

	if mimeType == "" {
		mimeType = fsimpl.ContentType(fi)
	}

	var data []byte

	if fi.IsDir() {
		var dirents []fs.DirEntry
		dirents, err = fs.ReadDir(fsys, fname)
		if err != nil {
			return nil, fmt.Errorf("readDir (url: %q, name: %s): %w", u, fname, err)
		}

		entries := make([]string, len(dirents))
		for i, e := range dirents {
			entries[i] = e.Name()
		}
		data, err = json.Marshal(entries)
		if err != nil {
			return nil, fmt.Errorf("json.Marshal: %w", err)
		}

		mimeType = iohelpers.JSONArrayMimetype
	} else {
		data, err = io.ReadAll(f)
		if err != nil {
			return nil, fmt.Errorf("read (url: %q, name: %s): %w", u, fname, err)
		}
	}

	if mimeType == "" {
		// default to text/plain
		mimeType = iohelpers.TextMimetype
	}

	return &content{contentType: mimeType, b: data}, nil
}

// resolveURL parses the relative URL rel against base, and returns the
// resolved URL. Differs from url.ResolveReference in that query parameters are
// added. In case of duplicates, params from rel are used.
func resolveURL(base url.URL, rel string) (*url.URL, error) {
	if rel == "" {
		return &base, nil
	}

	// git URLs are special - they have double-slashes that separate a repo
	// from a path in the repo. A missing double-slash means the path is the
	// root.
	switch base.Scheme {
	case "git", "git+file", "git+http", "git+https", "git+ssh":
		if strings.Contains(base.Path, "//") && strings.Contains(rel, "//") {
			return nil, fmt.Errorf("both base URL and subpath contain '//', which is not allowed in git URLs")
		}

		// If there's a subpath, the base path must end with '/'. This behaviour
		// is unique to git URLs - other schemes would instead drop the last
		// path element and replace with the subpath.
		if rel != "" && !strings.HasSuffix(base.Path, "/") {
			base.Path += "/"
		}

		// If subpath starts with '//', make it relative by prefixing a '.',
		// otherwise it'll be treated as a schemeless URI and the first part
		// will be interpreted as a hostname.
		if strings.HasPrefix(rel, "//") {
			rel = "." + rel
		}
	case "aws+sm":
		// aws+sm URLs may be opaque, so resolution needs to be handled
		// differently
		if base.Opaque != "" {
			// if it's opaque and we have a relative path we'll append it to
			// the opaque part
			if rel != "" {
				base.Opaque = path.Join(base.Opaque, rel)
			}

			return &base, nil
		} else if base.Path == "" && !strings.HasPrefix(rel, "/") {
			// if the base has no path and the relative URL doesn't start with
			// a slash, we treat it as opaque
			base.Opaque = rel
		}
	}

	// if there's still an opaque part, there's no resolving to do - just return
	// the base URL
	if base.Opaque != "" {
		return &base, nil
	}

	relURL, err := url.Parse(rel)
	if err != nil {
		return nil, err
	}

	// URL.ResolveReference requires (or assumes, at least) that the base is
	// absolute. We want to support relative URLs too though, so we need to
	// correct for that.
	var out *url.URL
	switch {
	case base.IsAbs():
		out = base.ResolveReference(relURL)
	case base.Scheme == "" && base.Path[0] == '/':
		// absolute path, no scheme or volume
		out = base.ResolveReference(relURL)
		out.Path = out.Path[1:]
	default:
		out = resolveRelativeURL(&base, relURL)
	}

	if base.RawQuery != "" {
		bq := base.Query()
		rq := relURL.Query()
		for k := range rq {
			bq.Set(k, rq.Get(k))
		}
		out.RawQuery = bq.Encode()
	}

	return out, nil
}

// relative path, might even involve .. or . in the path
func resolveRelativeURL(base, relURL *url.URL) *url.URL {
	// first find the difference between base and what base would be if it
	// were absolute
	emptyURL, _ := url.Parse("")
	absBase := base.ResolveReference(emptyURL)
	absBase.Path = strings.TrimPrefix(absBase.Path, "/")

	diff := strings.TrimSuffix(base.Path, absBase.Path)
	diff = strings.TrimSuffix(diff, "/")

	out := base.ResolveReference(relURL)

	// now correct the path by adding the prefix back in
	out.Path = diff + out.Path

	return out
}
