package net

import (
	"log"
	"net"
)

// LookupIP -
func LookupIP(name string) string {
	i := LookupIPs(name)
	if len(i) == 0 {
		return ""
	}
	return i[0]
}

// LookupIPs -
func LookupIPs(name string) []string {
	srcIPs, err := net.LookupIP(name)
	if err != nil {
		log.Fatal(err)
	}
	var ips []string
	for _, v := range srcIPs {
		if v.To4() != nil {
			ips = append(ips, v.String())
		}
	}
	return ips
}

// LookupCNAME -
func LookupCNAME(name string) string {
	cname, err := net.LookupCNAME(name)
	if err != nil {
		log.Fatal(err)
	}
	return cname
}

// LookupTXT -
func LookupTXT(name string) []string {
	records, err := net.LookupTXT(name)
	if err != nil {
		log.Fatal(err)
	}
	return records
}

// LookupSRV -
func LookupSRV(name string) *net.SRV {
	return LookupSRVs(name)[0]
}

// LookupSRVs -
func LookupSRVs(name string) []*net.SRV {
	_, addrs, err := net.LookupSRV("", "", name)
	if err != nil {
		log.Fatal(err)
	}
	return addrs
}
