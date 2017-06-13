package funcs

import (
	stdnet "net"
	"sync"

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
func (f *NetFuncs) LookupIP(name string) string {
	return net.LookupIP(name)
}

// LookupIPs -
func (f *NetFuncs) LookupIPs(name string) []string {
	return net.LookupIPs(name)
}

// LookupCNAME -
func (f *NetFuncs) LookupCNAME(name string) string {
	return net.LookupCNAME(name)
}

// LookupSRV -
func (f *NetFuncs) LookupSRV(name string) *stdnet.SRV {
	return net.LookupSRV(name)
}

// LookupSRVs -
func (f *NetFuncs) LookupSRVs(name string) []*stdnet.SRV {
	return net.LookupSRVs(name)
}

// LookupTXT -
func (f *NetFuncs) LookupTXT(name string) []string {
	return net.LookupTXT(name)
}
