// Package net contains functions to help with network-oriented lookups
package net

import (
	"net"
	"slices"
)

// LookupIP -
func LookupIP(name string) (string, error) {
	i, err := LookupIPs(name)
	if err != nil {
		return "", err
	}
	if len(i) == 0 {
		return "", nil
	}
	return i[0], nil
}

// LookupIPs -
func LookupIPs(name string) ([]string, error) {
	srcIPs, err := net.LookupIP(name)
	if err != nil {
		return nil, err
	}

	// perf note: this slice is not really worth pre-allocating - srcIPs tends
	// to be very small, and net.LookupIP is relatively expensive
	var ips []string
	for _, v := range srcIPs {
		if v.To4() != nil {
			s := v.String()
			if !slices.Contains(ips, s) {
				ips = append(ips, s)
			}
		}
	}
	return ips, nil
}

// LookupCNAME -
func LookupCNAME(name string) (string, error) {
	return net.LookupCNAME(name)
}

// LookupTXT -
func LookupTXT(name string) ([]string, error) {
	return net.LookupTXT(name)
}

// LookupSRV -
func LookupSRV(name string) (*net.SRV, error) {
	srvs, err := LookupSRVs(name)
	if err != nil {
		return nil, err
	}
	return srvs[0], nil
}

// LookupSRVs -
func LookupSRVs(name string) ([]*net.SRV, error) {
	_, addrs, err := net.LookupSRV("", "", name)
	if err != nil {
		return nil, err
	}
	return addrs, nil
}
