package funcs

import (
	"context"

	"github.com/hashicorp/go-sockaddr"
	"github.com/hashicorp/go-sockaddr/template"
)

// CreateSockaddrFuncs -
func CreateSockaddrFuncs(ctx context.Context) map[string]any {
	ns := &SockaddrFuncs{ctx}
	return map[string]any{
		"sockaddr": func() any { return ns },
	}
}

// SockaddrFuncs -
type SockaddrFuncs struct {
	ctx context.Context
}

// GetAllInterfaces -
func (SockaddrFuncs) GetAllInterfaces() (sockaddr.IfAddrs, error) {
	return sockaddr.GetAllInterfaces()
}

// GetDefaultInterfaces -
func (SockaddrFuncs) GetDefaultInterfaces() (sockaddr.IfAddrs, error) {
	return sockaddr.GetDefaultInterfaces()
}

// GetPrivateInterfaces -
func (SockaddrFuncs) GetPrivateInterfaces() (sockaddr.IfAddrs, error) {
	return sockaddr.GetPrivateInterfaces()
}

// GetPublicInterfaces -
func (SockaddrFuncs) GetPublicInterfaces() (sockaddr.IfAddrs, error) {
	return sockaddr.GetPublicInterfaces()
}

// Sort -
func (SockaddrFuncs) Sort(selectorParam string, inputIfAddrs sockaddr.IfAddrs) (sockaddr.IfAddrs, error) {
	return sockaddr.SortIfBy(selectorParam, inputIfAddrs)
}

// Exclude -
func (SockaddrFuncs) Exclude(selectorName, selectorParam string, inputIfAddrs sockaddr.IfAddrs) (sockaddr.IfAddrs, error) {
	return sockaddr.ExcludeIfs(selectorName, selectorParam, inputIfAddrs)
}

// Include -
func (SockaddrFuncs) Include(selectorName, selectorParam string, inputIfAddrs sockaddr.IfAddrs) (sockaddr.IfAddrs, error) {
	return sockaddr.IncludeIfs(selectorName, selectorParam, inputIfAddrs)
}

// Attr -
func (SockaddrFuncs) Attr(selectorName string, ifAddrsRaw any) (string, error) {
	return template.Attr(selectorName, ifAddrsRaw)
}

// Join -
func (SockaddrFuncs) Join(selectorName, joinString string, inputIfAddrs sockaddr.IfAddrs) (string, error) {
	return sockaddr.JoinIfAddrs(selectorName, joinString, inputIfAddrs)
}

// Limit -
func (SockaddrFuncs) Limit(lim uint, in sockaddr.IfAddrs) (sockaddr.IfAddrs, error) {
	return sockaddr.LimitIfAddrs(lim, in)
}

// Offset -
func (SockaddrFuncs) Offset(off int, in sockaddr.IfAddrs) (sockaddr.IfAddrs, error) {
	return sockaddr.OffsetIfAddrs(off, in)
}

// Unique -
func (SockaddrFuncs) Unique(selectorName string, inputIfAddrs sockaddr.IfAddrs) (sockaddr.IfAddrs, error) {
	return sockaddr.UniqueIfAddrsBy(selectorName, inputIfAddrs)
}

// Math -
func (SockaddrFuncs) Math(operation, value string, inputIfAddrs sockaddr.IfAddrs) (sockaddr.IfAddrs, error) {
	return sockaddr.IfAddrsMath(operation, value, inputIfAddrs)
}

// GetPrivateIP -
func (SockaddrFuncs) GetPrivateIP() (string, error) {
	return sockaddr.GetPrivateIP()
}

// GetPrivateIPs -
func (SockaddrFuncs) GetPrivateIPs() (string, error) {
	return sockaddr.GetPrivateIPs()
}

// GetPublicIP -
func (SockaddrFuncs) GetPublicIP() (string, error) {
	return sockaddr.GetPublicIP()
}

// GetPublicIPs -
func (SockaddrFuncs) GetPublicIPs() (string, error) {
	return sockaddr.GetPublicIPs()
}

// GetInterfaceIP -
func (SockaddrFuncs) GetInterfaceIP(namedIfRE string) (string, error) {
	return sockaddr.GetInterfaceIP(namedIfRE)
}

// GetInterfaceIPs -
func (SockaddrFuncs) GetInterfaceIPs(namedIfRE string) (string, error) {
	return sockaddr.GetInterfaceIPs(namedIfRE)
}
