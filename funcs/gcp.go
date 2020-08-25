package funcs

import (
	"context"
	"sync"

	"github.com/hairyhenderson/gomplate/v3/gcp"
)

var (
	gcpf     *GcpFuncs
	gcpfInit sync.Once
)

// GCPNS - the gcp namespace
func GCPNS() *GcpFuncs {
	gcpfInit.Do(func() {
		gcpf = &GcpFuncs{
			gcpopts: gcp.GetClientOptions(),
		}
	})
	return gcpf
}

// AddGCPFuncs -
// Deprecated: use CreateGCPFuncs
func AddGCPFuncs(f map[string]interface{}) {
	for k, v := range CreateGCPFuncs(context.Background()) {
		f[k] = v
	}
}

// CreateGCPFuncs -
func CreateGCPFuncs(ctx context.Context) map[string]interface{} {
	f := map[string]interface{}{}
	ns := GCPNS()
	ns.ctx = ctx
	f["gcp"] = GCPNS
	return f
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
