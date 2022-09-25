package funcs

import (
	"context"
	"sync"

	"github.com/hairyhenderson/gomplate/v3/gcp"
)

// GCPNS - the gcp namespace
//
// Deprecated: don't use
func GCPNS() *GcpFuncs {
	return &GcpFuncs{gcpopts: gcp.GetClientOptions()}
}

// AddGCPFuncs -
//
// Deprecated: use [CreateGCPFuncs] instead
func AddGCPFuncs(f map[string]interface{}) {
	for k, v := range CreateGCPFuncs(context.Background()) {
		f[k] = v
	}
}

// CreateGCPFuncs -
func CreateGCPFuncs(ctx context.Context) map[string]interface{} {
	ns := &GcpFuncs{
		ctx:     ctx,
		gcpopts: gcp.GetClientOptions(),
	}
	return map[string]interface{}{
		"gcp": func() interface{} { return ns },
	}
}

// GcpFuncs -
type GcpFuncs struct {
	ctx context.Context

	meta     *gcp.MetaClient
	metaInit sync.Once
	gcpopts  gcp.ClientOptions
}

// Meta -
func (a *GcpFuncs) Meta(key string, def ...string) (string, error) {
	a.metaInit.Do(a.initGcpMeta)
	return a.meta.Meta(key, def...)
}

func (a *GcpFuncs) initGcpMeta() {
	if a.meta == nil {
		a.meta = gcp.NewMetaClient(a.gcpopts)
	}
}
