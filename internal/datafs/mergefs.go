package datafs

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/hairyhenderson/go-fsimpl"
	"github.com/hairyhenderson/gomplate/v4/coll"
	"github.com/hairyhenderson/gomplate/v4/internal/config"
	"github.com/hairyhenderson/gomplate/v4/internal/iohelpers"
	"github.com/hairyhenderson/gomplate/v4/internal/parsers"
	"github.com/hairyhenderson/gomplate/v4/internal/urlhelpers"
)

// newMergeFS returns a new filesystem that merges the contents of multiple
// paths. Only a URL like "merge:" or "merge:///" makes sense here - the
// piped-separated lists of sub-sources to merge must be given to Open.
//
// You can use WithDataSourceRegistryFS to provide the datasource registry,
// otherwise, an empty registry will be used.
//
// An FSProvider will also be needed, which can be provided with a context
// using ContextWithFSProvider. Provide that context with fsimpl.WithContextFS.
func newMergeFS(u *url.URL) (fs.FS, error) {
	if u.Scheme != "merge" {
		return nil, fmt.Errorf("unsupported scheme %q", u.Scheme)
	}

	return &mergeFS{
		ctx:      context.Background(),
		registry: NewRegistry(),
	}, nil
}

type mergeFS struct {
	ctx        context.Context
	httpClient *http.Client
	registry   Registry
}

//nolint:gochecknoglobals
var mergeFSProvider = fsimpl.FSProviderFunc(newMergeFS, "merge")

var (
	_ fs.FS                    = (*mergeFS)(nil)
	_ withContexter            = (*mergeFS)(nil)
	_ withDataSourceRegistryer = (*mergeFS)(nil)
)

func (f *mergeFS) WithContext(ctx context.Context) fs.FS {
	if ctx == nil {
		return f
	}

	fsys := *f
	fsys.ctx = ctx

	return &fsys
}

func (f *mergeFS) WithHTTPClient(client *http.Client) fs.FS {
	if client == nil {
		return f
	}

	fsys := *f
	fsys.httpClient = client

	return &fsys
}

func (f *mergeFS) WithDataSourceRegistry(registry Registry) fs.FS {
	if registry == nil {
		return f
	}

	fsys := *f
	fsys.registry = registry

	return &fsys
}

func (f *mergeFS) Open(name string) (fs.File, error) {
	parts := strings.Split(name, "|")
	if len(parts) < 2 {
		return nil, &fs.PathError{
			Op: "open", Path: name,
			Err: fmt.Errorf("need at least 2 datasources to merge"),
		}
	}

	// now open each of the sub-files
	subFiles := make([]subFile, len(parts))

	modTime := time.Time{}

	for i, part := range parts {
		// if this is a datasource, look it up
		subSource, ok := f.registry.Lookup(part)
		if !ok {
			// maybe it's a relative filename?
			u, uerr := urlhelpers.ParseSourceURL(part)
			if uerr != nil {
				return nil, fmt.Errorf("unknown datasource %q, and couldn't parse URL: %w", part, uerr)
			}
			subSource = config.DataSource{URL: u}
		}

		u := subSource.URL

		// possible type hint in the type query param. Contrary to spec, we allow
		// unescaped '+' characters to make it simpler to provide types like
		// "application/array+json"
		overrideType := typeOverrideParam()
		mimeTypeHint := u.Query().Get(overrideType)
		mimeTypeHint = strings.ReplaceAll(mimeTypeHint, " ", "+")

		// now that we have the hint, remove it from the URL - we can't have it
		// leaking into the filesystem layer
		u = removeQueryParam(u, overrideType)

		fsURL, base := SplitFSMuxURL(u)

		// need to support absolute paths on local filesystem too
		// TODO: this is a hack, probably fix this?
		if fsURL.Scheme == "file" && runtime.GOOS != osWindows {
			base = fsURL.Path + base
		}

		fsys, err := FSysForPath(f.ctx, fsURL.String())
		if err != nil {
			return nil, &fs.PathError{
				Op: "open", Path: name,
				Err: fmt.Errorf("lookup for %s: %w", u.String(), err),
			}
		}

		// pass in the context and other bits
		fsys = fsimpl.WithContextFS(f.ctx, fsys)
		fsys = fsimpl.WithHeaderFS(subSource.Header, fsys)

		fsys = fsimpl.WithHTTPClientFS(f.httpClient, fsys)

		f, err := fsys.Open(base)
		if err != nil {
			return nil, &fs.PathError{
				Op: "open", Path: name,
				Err: fmt.Errorf("opening merge part %q: %w", part, err),
			}
		}

		subFiles[i] = subFile{f, mimeTypeHint}
	}

	return &mergeFile{
		name:     name,
		subFiles: subFiles,
		modTime:  modTime,
	}, nil
}

type subFile struct {
	fs.File
	contentType string
}

type mergeFile struct {
	name     string
	merged   io.Reader // the file's contents, post-merge - buffered here to enable partial reads
	fi       fs.FileInfo
	modTime  time.Time // the modTime of the most recently modified sub-file
	subFiles []subFile
	readMux  sync.Mutex
}

var _ fs.File = (*mergeFile)(nil)

func (f *mergeFile) Close() error {
	for _, f := range f.subFiles {
		f.Close()
	}
	return nil
}

func (f *mergeFile) Stat() (fs.FileInfo, error) {
	if f.merged == nil {
		p := make([]byte, 0)
		_, err := f.Read(p)
		if err != nil && !errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("read: %w", err)
		}
	}

	return f.fi, nil
}

func (f *mergeFile) Read(p []byte) (int, error) {
	// read from all and merge, then return the requested amount
	if f.merged == nil {
		f.readMux.Lock()
		defer f.readMux.Unlock()

		// read from all and merge
		data := make([]map[string]any, len(f.subFiles))
		for i, sf := range f.subFiles {
			d, err := f.readSubFile(sf)
			if err != nil {
				return 0, fmt.Errorf("readSubFile: %w", err)
			}

			data[i] = d
		}

		md, err := mergeData(data)
		if err != nil {
			return 0, fmt.Errorf("mergeData: %w", err)
		}

		f.merged = bytes.NewReader(md)

		f.fi = FileInfo(f.name, int64(len(md)), 0o400, f.modTime, iohelpers.YAMLMimetype)
	}

	return f.merged.Read(p)
}

func (f *mergeFile) readSubFile(sf subFile) (map[string]any, error) {
	// stat for content type and modTime
	fi, err := sf.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat merge part %q: %w", f.name, err)
	}

	// the merged file's modTime is the most recent of all the sub-files
	if fi.ModTime().After(f.modTime) {
		f.modTime = fi.ModTime()
	}

	// if we haven't been given a content type hint, guess the normal way
	if sf.contentType == "" {
		sf.contentType = fsimpl.ContentType(fi)
	}

	b, err := io.ReadAll(sf)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("readAll: %w", err)
	}

	sfData, err := parseMap(sf.contentType, string(b))
	if err != nil {
		return nil, fmt.Errorf("parsing map with content type %s: %w", sf.contentType, err)
	}

	return sfData, nil
}

func mergeData(data []map[string]any) ([]byte, error) {
	dst := data[0]
	data = data[1:]

	dst, err := coll.Merge(dst, data...)
	if err != nil {
		return nil, err
	}

	s, err := parsers.ToYAML(dst)
	if err != nil {
		return nil, err
	}
	return []byte(s), nil
}

func parseMap(mimeType, data string) (map[string]any, error) {
	datum, err := parsers.ParseData(mimeType, data)
	if err != nil {
		return nil, fmt.Errorf("parseData: %w", err)
	}

	var m map[string]any
	switch datum := datum.(type) {
	case map[string]any:
		m = datum
	default:
		return nil, fmt.Errorf("unexpected data type '%T' for datasource (type %s); merge: can only merge maps", datum, mimeType)
	}

	return m, nil
}
