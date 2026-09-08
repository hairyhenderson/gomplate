// Package net contains functions to help with network-oriented lookups
package net

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"slices"
	"strings"
)

const macLen = 6

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

// GenerateMAC generates a 6-octet MAC (hardware) address, returned as a string
// in the usual colon-separated form.
//
// prefix, when non-empty, fixes the leading octets of the address (for example
// an OUI like "aa:bb:cc"); the remaining octets are filled in randomly. Colons,
// hyphens, and dots in prefix are ignored, so "aa:bb", "aa-bb", and "aabb" are
// all equivalent. When prefix is empty the whole address is generated, and the
// local bit is set (and the multicast bit cleared) so the result is a locally
// administered unicast address that won't collide with real vendor-assigned
// hardware.
//
// When seed is non-nil the random octets are derived from it, so the same
// prefix and seed always produce the same address. When seed is nil the address
// is cryptographically random.
func GenerateMAC(prefix string, seed []byte) (string, error) {
	octets, err := parseMACPrefix(prefix)
	if err != nil {
		return "", err
	}

	mac := make(net.HardwareAddr, macLen)
	copy(mac, octets)
	fill := mac[len(octets):]

	if seed == nil {
		if _, err := rand.Read(fill); err != nil {
			return "", fmt.Errorf("failed to read random bytes: %w", err)
		}
	} else {
		sum := sha256.Sum256(seed)
		copy(fill, sum[:])
	}

	// With no prefix the first octet is generated too, so make the address a
	// locally administered unicast one to keep it out of real vendor ranges.
	if len(octets) == 0 {
		mac[0] = (mac[0] &^ 0x01) | 0x02
	}

	return mac.String(), nil
}

func parseMACPrefix(prefix string) ([]byte, error) {
	cleaned := strings.NewReplacer(":", "", "-", "", ".", "").Replace(prefix)
	if cleaned == "" {
		return nil, nil
	}

	octets, err := hex.DecodeString(cleaned)
	if err != nil {
		return nil, fmt.Errorf("invalid MAC prefix %q: %w", prefix, err)
	}
	if len(octets) > macLen {
		return nil, fmt.Errorf("MAC prefix %q is too long: want at most %d octets, got %d", prefix, macLen, len(octets))
	}

	return octets, nil
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
