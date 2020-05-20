package datasource

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

var headerKey = struct{}{}

// Source can read data from a specific data source
type Source interface {
	Read(ctx context.Context, args ...string) (*Data, error)
	Cleanup()
}

// SourceRegistry -
type SourceRegistry interface {
	// Register - create a new Source, appropriate for the url's scheme.
	// Data read by this source will be cached, and further calls with the same
	// args will hit the cache.
	Register(alias string, url *url.URL, header http.Header) (Source, error)
	// Exists reports whether or not a source with the given alias exists
	Exists(alias string) bool
	// Get returns a cached source if it exists
	Get(alias string) Source
	// Dynamic registers a new dynamically-defined source - the alias would be a URL in this case
	Dynamic(alias string, header http.Header) (Source, error)
	// ResetSources clears the cache of Sources. Should usually only be used for
	// testing.
	// Reset()
}

// DefaultRegistry - the default SourceRegistry
var DefaultRegistry SourceRegistry = &srcRegistry{
	sources: map[string]Source{},
}

// src - a Source implementation. Reuses a Reader. Uses a global cache to
// cache results.
type src struct {
	alias  string
	url    *url.URL
	header http.Header
	r      Reader
}

func (s *src) cacheKey(args ...string) string {
	key := s.alias
	for _, a := range args {
		key += a
	}
	return key
}

var dataCache = map[string]*Data{}

// Read -
func (s *src) Read(ctx context.Context, args ...string) (*Data, error) {
	if cached, ok := dataCache[s.cacheKey(args...)]; ok {
		return cached, nil
	}
	ctx = context.WithValue(ctx, headerKey, s.header)
	data, err := s.r.Read(ctx, s.url, args...)
	if err == nil && data != nil {
		dataCache[s.cacheKey(args...)] = data
	}
	return data, err
}

func (s *src) Cleanup() {
	if c, ok := s.r.(cleanerupper); ok {
		c.cleanup()
	}
}

// String is the method to format the flag's value, part of the flag.Value interface.
// The String method's output will be used in diagnostics.
func (s *src) String() string {
	return fmt.Sprintf("%s=%s", s.alias, s.url.String())
}

// srcRegistry - the default SourceRegistry
type srcRegistry struct {
	// sources - map of registered sources, indexed by alias
	sources map[string]Source
}

var _ SourceRegistry = (*srcRegistry)(nil)

// NewSource - create a new Source, appropriate for the url's scheme.
// Data read by this source will be cached, and further calls with the same
// args will hit the cache.
func (r *srcRegistry) Register(alias string, url *url.URL, header http.Header) (Source, error) {
	if readers == nil {
		initReaders()
	}

	reader, ok := readers[url.Scheme]
	if !ok {
		return nil, fmt.Errorf("scheme %s not registered", url.Scheme)
	}

	s := &src{
		alias:  alias,
		url:    url,
		header: header,
		r:      reader,
	}
	r.sources[alias] = s
	return s, nil
}

// Exists reports whether or not a source with the given alias exists
func (r *srcRegistry) Exists(alias string) bool {
	_, ok := r.sources[alias]
	return ok
}

// Get returns a cached source if it exists
func (r *srcRegistry) Get(alias string) Source {
	return r.sources[alias]
}

// Dynamic creates a new dynamically-defined source - the alias would be a URL in this case
func (r *srcRegistry) Dynamic(alias string, header http.Header) (Source, error) {
	srcURL, err := url.Parse(alias)
	if err != nil || !srcURL.IsAbs() {
		return nil, errors.Errorf("Undefined datasource '%s'", alias)
	}
	return r.Register(alias, srcURL, header)
}

// ResetSources clears the cache of Sources. Should usually only be used for
// testing.
func (r *srcRegistry) Reset() {
	r.sources = map[string]Source{}
}
