package net

import (
	"net"
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
	var ips []string
	for _, v := range srcIPs {
		if v.To4() != nil && !contains(ips, v.String()) {
			ips = append(ips, v.String())
		}
	}
	return ips, nil
}

func contains(a []string, s string) bool {
	for _, v := range a {
		if v == s {
			return true
		}
	}
	return false
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
