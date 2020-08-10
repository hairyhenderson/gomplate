package funcs

import (
	"context"
	"sync"

	"github.com/hashicorp/go-sockaddr"
	"github.com/hashicorp/go-sockaddr/template"
)

var (
	sockaddrNS     *SockaddrFuncs
	sockaddrNSInit sync.Once
)

// SockaddrNS - the sockaddr namespace
func SockaddrNS() *SockaddrFuncs {
	sockaddrNSInit.Do(func() { sockaddrNS = &SockaddrFuncs{} })
	return sockaddrNS
}

// AddSockaddrFuncs -
func AddSockaddrFuncs(f map[string]interface{}) {
	f["sockaddr"] = SockaddrNS
}

// CreateSockaddrFuncs -
func CreateSockaddrFuncs(ctx context.Context) map[string]interface{} {
	ns := SockaddrNS()
	ns.ctx = ctx
	return map[string]interface{}{"sockaddr": SockaddrNS}
}

// SockaddrFuncs -
type SockaddrFuncs struct {
	ctx context.Context
}

// GetAllInterfaces -
func (f *SockaddrFuncs) GetAllInterfaces() (sockaddr.IfAddrs, error) {
	return sockaddr.GetAllInterfaces()
}

// GetDefaultInterfaces -
func (f *SockaddrFuncs) GetDefaultInterfaces() (sockaddr.IfAddrs, error) {
	return sockaddr.GetDefaultInterfaces()
}

// GetPrivateInterfaces -
func (f *SockaddrFuncs) GetPrivateInterfaces() (sockaddr.IfAddrs, error) {
	return sockaddr.GetPrivateInterfaces()
}

// GetPublicInterfaces -
func (f *SockaddrFuncs) GetPublicInterfaces() (sockaddr.IfAddrs, error) {
	return sockaddr.GetPublicInterfaces()
}

// Sort -
func (f *SockaddrFuncs) Sort(selectorParam string, inputIfAddrs sockaddr.IfAddrs) (sockaddr.IfAddrs, error) {
	return sockaddr.SortIfBy(selectorParam, inputIfAddrs)
}

// Exclude -
func (f *SockaddrFuncs) Exclude(selectorName, selectorParam string, inputIfAddrs sockaddr.IfAddrs) (sockaddr.IfAddrs, error) {
	return sockaddr.ExcludeIfs(selectorName, selectorParam, inputIfAddrs)
}

// Include -
func (f *SockaddrFuncs) Include(selectorName, selectorParam string, inputIfAddrs sockaddr.IfAddrs) (sockaddr.IfAddrs, error) {
	return sockaddr.IncludeIfs(selectorName, selectorParam, inputIfAddrs)
}

// Attr -
func (f *SockaddrFuncs) Attr(selectorName string, ifAddrsRaw interface{}) (string, error) {
	return template.Attr(selectorName, ifAddrsRaw)
}

// Join -
func (f *SockaddrFuncs) Join(selectorName, joinString string, inputIfAddrs sockaddr.IfAddrs) (string, error) {
	return sockaddr.JoinIfAddrs(selectorName, joinString, inputIfAddrs)
}

// Limit -
func (f *SockaddrFuncs) Limit(lim uint, in sockaddr.IfAddrs) (sockaddr.IfAddrs, error) {
	return sockaddr.LimitIfAddrs(lim, in)
}

// Offset -
func (f *SockaddrFuncs) Offset(off int, in sockaddr.IfAddrs) (sockaddr.IfAddrs, error) {
	return sockaddr.OffsetIfAddrs(off, in)
}

// Unique -
func (f *SockaddrFuncs) Unique(selectorName string, inputIfAddrs sockaddr.IfAddrs) (sockaddr.IfAddrs, error) {
	return sockaddr.UniqueIfAddrsBy(selectorName, inputIfAddrs)
}

// Math -
func (f *SockaddrFuncs) Math(operation, value string, inputIfAddrs sockaddr.IfAddrs) (sockaddr.IfAddrs, error) {
	return sockaddr.IfAddrsMath(operation, value, inputIfAddrs)
}

// GetPrivateIP -
func (f *SockaddrFuncs) GetPrivateIP() (string, error) {
	return sockaddr.GetPrivateIP()
}

// GetPrivateIPs -
func (f *SockaddrFuncs) GetPrivateIPs() (string, error) {
	return sockaddr.GetPrivateIPs()
}

// GetPublicIP -
func (f *SockaddrFuncs) GetPublicIP() (string, error) {
	return sockaddr.GetPublicIP()
}

// GetPublicIPs -
func (f *SockaddrFuncs) GetPublicIPs() (string, error) {
	return sockaddr.GetPublicIPs()
}

// GetInterfaceIP -
func (f *SockaddrFuncs) GetInterfaceIP(namedIfRE string) (string, error) {
	return sockaddr.GetInterfaceIP(namedIfRE)
}

// GetInterfaceIPs -
func (f *SockaddrFuncs) GetInterfaceIPs(namedIfRE string) (string, error) {
	return sockaddr.GetInterfaceIPs(namedIfRE)
}
