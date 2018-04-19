package funcs

import (
	stdnet "net"
	"sync"

	"github.com/hairyhenderson/gomplate/conv"

	"github.com/hairyhenderson/gomplate/net"
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

// NetFuncs -
type NetFuncs struct{}

// LookupIP -
func (f *NetFuncs) LookupIP(name interface{}) string {
	return net.LookupIP(conv.ToString(name))
}

// LookupIPs -
func (f *NetFuncs) LookupIPs(name interface{}) []string {
	return net.LookupIPs(conv.ToString(name))
}

// LookupCNAME -
func (f *NetFuncs) LookupCNAME(name interface{}) string {
	return net.LookupCNAME(conv.ToString(name))
}

// LookupSRV -
func (f *NetFuncs) LookupSRV(name interface{}) *stdnet.SRV {
	return net.LookupSRV(conv.ToString(name))
}

// LookupSRVs -
func (f *NetFuncs) LookupSRVs(name interface{}) []*stdnet.SRV {
	return net.LookupSRVs(conv.ToString(name))
}

// LookupTXT -
func (f *NetFuncs) LookupTXT(name interface{}) []string {
	return net.LookupTXT(conv.ToString(name))
}
