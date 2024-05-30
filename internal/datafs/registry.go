package datafs

import (
	"net/http"
	"sort"
	"sync"

	"github.com/hairyhenderson/gomplate/v4/internal/config"
)

// Registry - a registry of datasources
type Registry interface {
	// Register a datasource
	Register(alias string, ds config.DataSource)
	// Lookup a registered datasource
	Lookup(alias string) (config.DataSource, bool)
	// List registered datasource aliases
	List() []string

	// Add extra headers not attached to a pre-defined datasource. These can be
	// used by datasources registered at runtime.
	AddExtraHeader(alias string, hdr http.Header)
}

func NewRegistry() Registry {
	return &dsRegistry{
		RWMutex:      &sync.RWMutex{},
		m:            map[string]config.DataSource{},
		extraHeaders: map[string]http.Header{},
	}
}

type dsRegistry struct {
	*sync.RWMutex
	m            map[string]config.DataSource
	extraHeaders map[string]http.Header
}

// Register a datasource
func (r *dsRegistry) Register(alias string, ds config.DataSource) {
	r.Lock()
	defer r.Unlock()

	// if there's an extra header for this datasource, and the datasource
	// doesn't have a header, add it now
	if hdr, ok := r.extraHeaders[alias]; ok && ds.Header == nil {
		ds.Header = hdr
	}

	r.m[alias] = ds
}

// Lookup a registered datasource
func (r *dsRegistry) Lookup(alias string) (config.DataSource, bool) {
	r.RLock()
	defer r.RUnlock()

	ds, ok := r.m[alias]
	if !ok {
		return ds, ok
	}

	return ds, ok
}

// List registered datasource aliases
func (r *dsRegistry) List() []string {
	r.RLock()
	defer r.RUnlock()

	keys := make([]string, 0, len(r.m))
	for k := range r.m {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}

// AddExtraHeader adds extra headers not attached (yet) to a datasource. These will be added
// to the headers of any matching datasource when Lookup is called.
func (r *dsRegistry) AddExtraHeader(alias string, hdr http.Header) {
	r.Lock()
	defer r.Unlock()

	r.extraHeaders[alias] = hdr
}
