package funcs

import (
	"context"
	"sync"

	"github.com/hairyhenderson/gomplate/v4/gcp"
)

// CreateGCPFuncs -
func CreateGCPFuncs(ctx context.Context) map[string]any {
	ns := &GcpFuncs{
		ctx:     ctx,
		gcpopts: gcp.GetClientOptions(),
	}
	return map[string]any{
		"gcp": func() any { return ns },
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
	a.meta = sync.OnceValue(func() *gcp.MetaClient {
		return gcp.NewMetaClient(a.ctx, a.gcpopts)
	})()

	return a.meta.Meta(key, def...)
}
