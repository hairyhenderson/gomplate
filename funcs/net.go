package funcs

import (
	"context"
	stdnet "net"

	"github.com/hairyhenderson/gomplate/v3/conv"

	"github.com/hairyhenderson/gomplate/v3/net"
)

// NetNS - the net namespace
// Deprecated: don't use
func NetNS() *NetFuncs {
	return &NetFuncs{}
}

// AddNetFuncs -
// Deprecated: use CreateNetFuncs instead
func AddNetFuncs(f map[string]interface{}) {
	for k, v := range CreateNetFuncs(context.Background()) {
		f[k] = v
	}
}

// CreateNetFuncs -
func CreateNetFuncs(ctx context.Context) map[string]interface{} {
	ns := &NetFuncs{ctx}
	return map[string]interface{}{
		"net": func() interface{} { return ns },
	}
}

// NetFuncs -
type NetFuncs struct {
	ctx context.Context
}

// LookupIP -
func (f NetFuncs) LookupIP(name interface{}) (string, error) {
	return net.LookupIP(conv.ToString(name))
}

// LookupIPs -
func (f NetFuncs) LookupIPs(name interface{}) ([]string, error) {
	return net.LookupIPs(conv.ToString(name))
}

// LookupCNAME -
func (f NetFuncs) LookupCNAME(name interface{}) (string, error) {
	return net.LookupCNAME(conv.ToString(name))
}

// LookupSRV -
func (f NetFuncs) LookupSRV(name interface{}) (*stdnet.SRV, error) {
	return net.LookupSRV(conv.ToString(name))
}

// LookupSRVs -
func (f NetFuncs) LookupSRVs(name interface{}) ([]*stdnet.SRV, error) {
	return net.LookupSRVs(conv.ToString(name))
}

// LookupTXT -
func (f NetFuncs) LookupTXT(name interface{}) ([]string, error) {
	return net.LookupTXT(conv.ToString(name))
}
