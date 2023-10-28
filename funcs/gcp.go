package funcs

import (
	"context"
	"sync"

	"github.com/hairyhenderson/gomplate/v4/gcp"
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

	meta    *gcp.MetaClient
	gcpopts gcp.ClientOptions
}

// Meta -
func (a *GcpFuncs) Meta(key string, def ...string) (string, error) {
	a.meta = sync.OnceValue[*gcp.MetaClient](func() *gcp.MetaClient {
		return gcp.NewMetaClient(a.ctx, a.gcpopts)
	})()

	return a.meta.Meta(key, def...)
}
