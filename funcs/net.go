package funcs

import (
	"context"
	stdnet "net"
	"sync"

	"github.com/hairyhenderson/gomplate/v3/conv"

	"github.com/hairyhenderson/gomplate/v3/net"
)

var (
	netNS     *NetFuncs
	netNSInit sync.Once
)

// NetNS - the net namespace
func NetNS() *NetFuncs {
	netNSInit.Do(func() { netNS = &NetFuncs{} })
	return netNS
}

// AddNetFuncs -
func AddNetFuncs(f map[string]interface{}) {
	f["net"] = NetNS
}

// CreateNetFuncs -
func CreateNetFuncs(ctx context.Context) map[string]interface{} {
	ns := NetNS()
	ns.ctx = ctx
	return map[string]interface{}{"net": NetNS}
}

// NetFuncs -
type NetFuncs struct {
	ctx context.Context
}

// LookupIP -
func (f *NetFuncs) LookupIP(name interface{}) (string, error) {
	return net.LookupIP(conv.ToString(name))
}

// LookupIPs -
func (f *NetFuncs) LookupIPs(name interface{}) ([]string, error) {
	return net.LookupIPs(conv.ToString(name))
}

// LookupCNAME -
func (f *NetFuncs) LookupCNAME(name interface{}) (string, error) {
	return net.LookupCNAME(conv.ToString(name))
}

// LookupSRV -
func (f *NetFuncs) LookupSRV(name interface{}) (*stdnet.SRV, error) {
	return net.LookupSRV(conv.ToString(name))
}

// LookupSRVs -
func (f *NetFuncs) LookupSRVs(name interface{}) ([]*stdnet.SRV, error) {
	return net.LookupSRVs(conv.ToString(name))
}

// LookupTXT -
func (f *NetFuncs) LookupTXT(name interface{}) ([]string, error) {
	return net.LookupTXT(conv.ToString(name))
}
