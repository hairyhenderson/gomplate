package data

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"runtime"
	"sort"
	"strings"

	"github.com/hairyhenderson/go-fsimpl"
	"github.com/hairyhenderson/gomplate/v4/internal/config"
	"github.com/hairyhenderson/gomplate/v4/internal/datafs"
	"github.com/hairyhenderson/gomplate/v4/internal/parsers"
	"github.com/hairyhenderson/gomplate/v4/internal/urlhelpers"
)

// Data -
//
// Deprecated: will be replaced in future
type Data struct {
	Ctx context.Context

	// TODO: remove this before 4.0
	Sources map[string]*Source

	cache map[string]*fileContent

	// headers from the --datasource-header/-H option that don't reference datasources from the commandline
	ExtraHeaders map[string]http.Header
}

type fileContent struct {
	contentType string
	b           []byte
}

// Cleanup - clean up datasources before shutting the process down - things
// like Logging out happen here
func (d *Data) Cleanup() {
	for _, s := range d.Sources {
		s.cleanup()
	}
}

// NewData - constructor for Data
//
// Deprecated: will be replaced in future
func NewData(datasourceArgs, headerArgs []string) (*Data, error) {
	cfg := &config.Config{}
	err := cfg.ParseDataSourceFlags(datasourceArgs, nil, nil, headerArgs)
	if err != nil {
		return nil, err
	}
	data := FromConfig(context.Background(), cfg)
	return data, nil
}

// FromConfig - internal use only!
func FromConfig(ctx context.Context, cfg *config.Config) *Data {
	// XXX: This is temporary, and will be replaced with something a bit cleaner
	// when datasources are refactored
	ctx = datafs.ContextWithStdin(ctx, cfg.Stdin)

	sources := map[string]*Source{}
	for alias, d := range cfg.DataSources {
		sources[alias] = &Source{
			Alias:  alias,
			URL:    d.URL,
			Header: d.Header,
		}
	}
	for alias, d := range cfg.Context {
		sources[alias] = &Source{
			Alias:  alias,
			URL:    d.URL,
			Header: d.Header,
		}
	}
	return &Data{
		Ctx:          ctx,
		Sources:      sources,
		ExtraHeaders: cfg.ExtraHeaders,
	}
}

// Source - a data source
//
// Deprecated: will be replaced in future
type Source struct {
	Alias     string
	URL       *url.URL
	Header    http.Header // used for http[s]: URLs, nil otherwise
	mediaType string
}

// Deprecated: no-op
func (s *Source) cleanup() {
	// if s.kv != nil {
	// 	s.kv.Logout()
	// }
}

// String is the method to format the flag's value, part of the flag.Value interface.
// The String method's output will be used in diagnostics.
func (s *Source) String() string {
	return fmt.Sprintf("%s=%s (%s)", s.Alias, s.URL.String(), s.mediaType)
}

// DefineDatasource -
func (d *Data) DefineDatasource(alias, value string) (string, error) {
	if alias == "" {
		return "", fmt.Errorf("datasource alias must be provided")
	}
	if d.DatasourceExists(alias) {
		return "", nil
	}
	srcURL, err := urlhelpers.ParseSourceURL(value)
	if err != nil {
		return "", err
	}
	s := &Source{
		Alias:  alias,
		URL:    srcURL,
		Header: d.ExtraHeaders[alias],
	}
	if d.Sources == nil {
		d.Sources = make(map[string]*Source)
	}
	d.Sources[alias] = s
	return "", nil
}

// DatasourceExists -
func (d *Data) DatasourceExists(alias string) bool {
	_, ok := d.Sources[alias]
	return ok
}

func (d *Data) lookupSource(alias string) (*Source, error) {
	source, ok := d.Sources[alias]
	if !ok {
		srcURL, err := url.Parse(alias)
		if err != nil || !srcURL.IsAbs() {
			return nil, fmt.Errorf("undefined datasource '%s': %w", alias, err)
		}
		source = &Source{
			Alias:  alias,
			URL:    srcURL,
			Header: d.ExtraHeaders[alias],
		}
		d.Sources[alias] = source
	}
	if source.Alias == "" {
		source.Alias = alias
	}
	return source, nil
}

func (d *Data) readDataSource(ctx context.Context, alias string, args ...string) (*fileContent, error) {
	source, err := d.lookupSource(alias)
	if err != nil {
		return nil, err
	}
	fc, err := d.readSource(ctx, source, args...)
	if err != nil {
		return nil, fmt.Errorf("couldn't read datasource '%s': %w", alias, err)
	}

	return fc, nil
}

// Include -
func (d *Data) Include(alias string, args ...string) (string, error) {
	fc, err := d.readDataSource(d.Ctx, alias, args...)
	if err != nil {
		return "", err
	}

	return string(fc.b), err
}

// Datasource -
func (d *Data) Datasource(alias string, args ...string) (interface{}, error) {
	fc, err := d.readDataSource(d.Ctx, alias, args...)
	if err != nil {
		return nil, err
	}

	return parsers.ParseData(fc.contentType, string(fc.b))
}

// DatasourceReachable - Determines if the named datasource is reachable with
// the given arguments. Reads from the datasource, and discards the returned data.
func (d *Data) DatasourceReachable(alias string, args ...string) bool {
	source, ok := d.Sources[alias]
	if !ok {
		return false
	}
	_, err := d.readSource(d.Ctx, source, args...)
	return err == nil
}

// readSource returns the (possibly cached) data from the given source,
// as referenced by the given args
func (d *Data) readSource(ctx context.Context, source *Source, args ...string) (*fileContent, error) {
	if d.cache == nil {
		d.cache = make(map[string]*fileContent)
	}
	cacheKey := source.Alias
	for _, v := range args {
		cacheKey += v
	}
	cached, ok := d.cache[cacheKey]
	if ok {
		return cached, nil
	}

	arg := ""
	if len(args) > 0 {
		arg = args[0]
	}
	u, err := resolveURL(source.URL, arg)
	if err != nil {
		return nil, err
	}

	fc, err := d.readFileContent(ctx, u, source.Header)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", u, err)
	}
	d.cache[cacheKey] = fc
	return fc, nil
}

// readFileContent returns content from the given URL
func (d Data) readFileContent(ctx context.Context, u *url.URL, hdr http.Header) (*fileContent, error) {
	fsys, err := datafs.FSysForPath(ctx, u.String())
	if err != nil {
		return nil, fmt.Errorf("fsys for path %v: %w", u, err)
	}

	u, fname := datafs.SplitFSMuxURL(u)

	// need to support absolute paths on local filesystem too
	// TODO: this is a hack, probably fix this?
	if u.Scheme == "file" && runtime.GOOS != "windows" {
		fname = u.Path + fname
	}

	fsys = fsimpl.WithContextFS(ctx, fsys)
	fsys = fsimpl.WithHeaderFS(hdr, fsys)

	// convert d.Sources to a map[string]config.DataSources
	// TODO: remove this when d.Sources is removed
	ds := make(map[string]config.DataSource)
	for k, v := range d.Sources {
		ds[k] = config.DataSource{
			URL:    v.URL,
			Header: v.Header,
		}
	}

	fsys = datafs.WithDataSourcesFS(ds, fsys)

	f, err := fsys.Open(fname)
	if err != nil {
		return nil, fmt.Errorf("open (url: %q, name: %q): %w", u, fname, err)
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat (url: %q, name: %q): %w", u, fname, err)
	}

	// possible type hint in the type query param. Contrary to spec, we allow
	// unescaped '+' characters to make it simpler to provide types like
	// "application/array+json"
	mimeType := u.Query().Get("type")
	mimeType = strings.ReplaceAll(mimeType, " ", "+")

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

		mimeType = jsonArrayMimetype
	} else {
		data, err = io.ReadAll(f)
		if err != nil {
			return nil, fmt.Errorf("read (url: %q, name: %s): %w", u, fname, err)
		}
	}

	if mimeType == "" {
		// default to text/plain
		mimeType = textMimetype
	}

	return &fileContent{contentType: mimeType, b: data}, nil
}

// Show all datasources  -
func (d *Data) ListDatasources() []string {
	datasources := make([]string, 0, len(d.Sources))
	for source := range d.Sources {
		datasources = append(datasources, source)
	}
	sort.Strings(datasources)
	return datasources
}

// resolveURL parses the relative URL rel against base, and returns the
// resolved URL. Differs from url.ResolveReference in that query parameters are
// added. In case of duplicates, params from rel are used.
func resolveURL(base *url.URL, rel string) (*url.URL, error) {
	// if there's an opaque part, there's no resolving to do - just return the
	// base URL
	if base.Opaque != "" {
		return base, nil
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
	}

	relURL, err := url.Parse(rel)
	if err != nil {
		return nil, err
	}

	// URL.ResolveReference requires (or assumes, at least) that the base is
	// absolute. We want to support relative URLs too though, so we need to
	// correct for that.
	out := base.ResolveReference(relURL)
	if out.Scheme == "" && out.Path[0] == '/' {
		out.Path = out.Path[1:]
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
