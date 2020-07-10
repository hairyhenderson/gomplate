package funcs

import (
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
func AddGCPFuncs(f map[string]interface{}) {
	f["gcp"] = GCPNS
}

// Funcs -
type GcpFuncs struct {
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
