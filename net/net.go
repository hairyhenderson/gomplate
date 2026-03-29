// Package net contains functions to help with network-oriented lookups
package net

import (
	"context"
	"net"
	"slices"
)

// LookupIP -
func LookupIP(ctx context.Context, name string) (string, error) {
	i, err := LookupIPs(ctx, name)
	if err != nil {
		return "", err
	}
	if len(i) == 0 {
		return "", nil
	}
	return i[0], nil
}

// LookupIPs -
func LookupIPs(ctx context.Context, name string) ([]string, error) {
	resolver := &net.Resolver{}
	srcIPs, err := resolver.LookupIPAddr(ctx, name)
	if err != nil {
		return nil, err
	}

	// perf note: this slice is not really worth pre-allocating - srcIPs tends
	// to be very small, and LookupIPAddr is relatively expensive
	var ips []string
	for _, v := range srcIPs {
		if v.IP.To4() != nil {
			s := v.IP.String()
			if !slices.Contains(ips, s) {
				ips = append(ips, s)
			}
		}
	}
	return ips, nil
}

// LookupCNAME -
//
// Deprecated: use [net.Resolver.LookupCNAME] instead
func LookupCNAME(name string) (string, error) {
	resolver := &net.Resolver{}
	return resolver.LookupCNAME(context.Background(), name)
}

// LookupTXT -
//
// Deprecated: use [net.Resolver.LookupTXT] instead
func LookupTXT(name string) ([]string, error) {
	resolver := &net.Resolver{}
	return resolver.LookupTXT(context.Background(), name)
}

// LookupSRV -
//
// Deprecated: use [net.Resolver#LookupSRV] instead
func LookupSRV(name string) (*net.SRV, error) {
	srvs, err := LookupSRVs(name)
	if err != nil {
		return nil, err
	}
	return srvs[0], nil
}

// LookupSRVs -
//
// Deprecated: use [net.Resolver#LookupSRV] instead
func LookupSRVs(name string) ([]*net.SRV, error) {
	resolver := &net.Resolver{}
	_, addrs, err := resolver.LookupSRV(context.Background(), "", "", name)
	if err != nil {
		return nil, err
	}
	return addrs, nil
}
