package funcs

import (
	"context"
	"fmt"
	"math/big"
	stdnet "net"
	"net/netip"

	"github.com/hairyhenderson/gomplate/v4/conv"
	"github.com/hairyhenderson/gomplate/v4/internal/cidr"
	"github.com/hairyhenderson/gomplate/v4/internal/deprecated"
	"github.com/hairyhenderson/gomplate/v4/net"
	"go4.org/netipx"
	"inet.af/netaddr"
)

// CreateNetFuncs -
func CreateNetFuncs(ctx context.Context) map[string]interface{} {
	ns := &NetFuncs{ctx}
	return map[string]interface{}{
		"net": func() interface{} { return ns },
	}
}

// NetFuncs -
type NetFuncs struct {
	ctx context.Context
}

// LookupIP -
func (f NetFuncs) LookupIP(name interface{}) (string, error) {
	return net.LookupIP(conv.ToString(name))
}

// LookupIPs -
func (f NetFuncs) LookupIPs(name interface{}) ([]string, error) {
	return net.LookupIPs(conv.ToString(name))
}

// LookupCNAME -
func (f NetFuncs) LookupCNAME(name interface{}) (string, error) {
	return net.LookupCNAME(conv.ToString(name))
}

// LookupSRV -
func (f NetFuncs) LookupSRV(name interface{}) (*stdnet.SRV, error) {
	return net.LookupSRV(conv.ToString(name))
}

// LookupSRVs -
func (f NetFuncs) LookupSRVs(name interface{}) ([]*stdnet.SRV, error) {
	return net.LookupSRVs(conv.ToString(name))
}

// LookupTXT -
func (f NetFuncs) LookupTXT(name interface{}) ([]string, error) {
	return net.LookupTXT(conv.ToString(name))
}

// ParseIP -
//
// Deprecated: use [ParseAddr] instead
func (f *NetFuncs) ParseIP(ip interface{}) (netaddr.IP, error) {
	deprecated.WarnDeprecated(f.ctx, "net.ParseIP is deprecated - use net.ParseAddr instead")
	return netaddr.ParseIP(conv.ToString(ip))
}

// ParseIPPrefix -
//
// Deprecated: use [ParsePrefix] instead
func (f *NetFuncs) ParseIPPrefix(ipprefix interface{}) (netaddr.IPPrefix, error) {
	deprecated.WarnDeprecated(f.ctx, "net.ParseIPPrefix is deprecated - use net.ParsePrefix instead")
	return netaddr.ParseIPPrefix(conv.ToString(ipprefix))
}

// ParseIPRange -
//
// Deprecated: use [ParseRange] instead
func (f *NetFuncs) ParseIPRange(iprange interface{}) (netaddr.IPRange, error) {
	deprecated.WarnDeprecated(f.ctx, "net.ParseIPRange is deprecated - use net.ParseRange instead")
	return netaddr.ParseIPRange(conv.ToString(iprange))
}

// ParseAddr -
func (f NetFuncs) ParseAddr(ip interface{}) (netip.Addr, error) {
	return netip.ParseAddr(conv.ToString(ip))
}

// ParsePrefix -
func (f NetFuncs) ParsePrefix(ipprefix interface{}) (netip.Prefix, error) {
	return netip.ParsePrefix(conv.ToString(ipprefix))
}

// ParseRange -
//
// Experimental: this API may change in the future
func (f NetFuncs) ParseRange(iprange interface{}) (netipx.IPRange, error) {
	return netipx.ParseIPRange(conv.ToString(iprange))
}

func (f *NetFuncs) parseNetipPrefix(prefix interface{}) (netip.Prefix, error) {
	switch p := prefix.(type) {
	case *stdnet.IPNet:
		return f.ipPrefixFromIPNet(p), nil
	case netaddr.IPPrefix:
		deprecated.WarnDeprecated(f.ctx,
			"support for netaddr.IPPrefix is deprecated - use net.ParsePrefix to produce a netip.Prefix instead")
		return f.ipPrefixFromIPNet(p.Masked().IPNet()), nil
	case netip.Prefix:
		return p, nil
	default:
		return netip.ParsePrefix(conv.ToString(prefix))
	}
}

func (f NetFuncs) ipPrefixFromIPNet(n *stdnet.IPNet) netip.Prefix {
	ip, _ := netip.AddrFromSlice(n.IP)
	ones, _ := n.Mask.Size()
	return netip.PrefixFrom(ip, ones)
}

// CIDRHost -
// Experimental!
func (f *NetFuncs) CIDRHost(hostnum interface{}, prefix interface{}) (netip.Addr, error) {
	if err := checkExperimental(f.ctx); err != nil {
		return netip.Addr{}, err
	}

	network, err := f.parseNetipPrefix(prefix)
	if err != nil {
		return netip.Addr{}, err
	}

	n, err := conv.ToInt64(hostnum)
	if err != nil {
		return netip.Addr{}, fmt.Errorf("expected a number: %w", err)
	}

	ip, err := cidr.HostBig(network, big.NewInt(n))

	return ip, err
}

// CIDRNetmask -
// Experimental!
func (f *NetFuncs) CIDRNetmask(prefix interface{}) (netip.Addr, error) {
	if err := checkExperimental(f.ctx); err != nil {
		return netip.Addr{}, err
	}

	p, err := f.parseNetipPrefix(prefix)
	if err != nil {
		return netip.Addr{}, err
	}

	// fill an appropriately sized byte slice with as many 1s as prefix bits
	b := make([]byte, p.Addr().BitLen()/8)
	for i := 0; i < p.Bits(); i++ {
		//nolint:gosec // G115 is not applicable, the value was checked at parse
		// time
		b[i/8] |= 1 << uint(7-i%8)
	}

	m, ok := netip.AddrFromSlice(b)
	if !ok {
		return netip.Addr{}, fmt.Errorf("invalid netmask")
	}

	return m, nil
}

// CIDRSubnets -
// Experimental!
func (f *NetFuncs) CIDRSubnets(newbits interface{}, prefix interface{}) ([]netip.Prefix, error) {
	if err := checkExperimental(f.ctx); err != nil {
		return nil, err
	}

	network, err := f.parseNetipPrefix(prefix)
	if err != nil {
		return nil, err
	}

	nBits, err := conv.ToInt(newbits)
	if err != nil {
		return nil, fmt.Errorf("newbits must be a number: %w", err)
	}

	if nBits < 1 {
		return nil, fmt.Errorf("must extend prefix by at least one bit")
	}

	maxNetNum := int64(1 << uint64(nBits))
	retValues := make([]netip.Prefix, maxNetNum)
	for i := int64(0); i < maxNetNum; i++ {
		subnet, err := cidr.SubnetBig(network, nBits, big.NewInt(i))
		if err != nil {
			return nil, err
		}
		retValues[i] = subnet
	}

	return retValues, nil
}

// CIDRSubnetSizes -
// Experimental!
func (f *NetFuncs) CIDRSubnetSizes(args ...interface{}) ([]netip.Prefix, error) {
	if err := checkExperimental(f.ctx); err != nil {
		return nil, err
	}

	if len(args) < 2 {
		return nil, fmt.Errorf("wrong number of args: want 2 or more, got %d", len(args))
	}

	network, err := f.parseNetipPrefix(args[len(args)-1])
	if err != nil {
		return nil, err
	}

	newbits, err := conv.ToInts(args[:len(args)-1]...)
	if err != nil {
		return nil, fmt.Errorf("newbits must be numbers: %w", err)
	}

	startPrefixLen := network.Bits()
	firstLength := newbits[0]

	firstLength += startPrefixLen
	retValues := make([]netip.Prefix, len(newbits))

	current, _ := cidr.PreviousSubnet(network, firstLength)

	for i, length := range newbits {
		if length < 1 {
			return nil, fmt.Errorf("must extend prefix by at least one bit")
		}
		// For portability with 32-bit systems where the subnet number
		// will be a 32-bit int, we only allow extension of 32 bits in
		// one call even if we're running on a 64-bit machine.
		// (Of course, this is significant only for IPv6.)
		if length > 32 {
			return nil, fmt.Errorf("may not extend prefix by more than 32 bits")
		}

		length += startPrefixLen
		if length > network.Addr().BitLen() {
			protocol := "IP"
			switch {
			case network.Addr().Is4():
				protocol = "IPv4"
			case network.Addr().Is6():
				protocol = "IPv6"
			}
			return nil, fmt.Errorf("would extend prefix to %d bits, which is too long for an %s address", length, protocol)
		}

		next, rollover := cidr.NextSubnet(current, length)
		if rollover || !network.Contains(next.Addr()) {
			// If we run out of suffix bits in the base CIDR prefix then
			// NextSubnet will start incrementing the prefix bits, which
			// we don't allow because it would then allocate addresses
			// outside of the caller's given prefix.
			return nil, fmt.Errorf("not enough remaining address space for a subnet with a prefix of %d bits after %s", length, current.String())
		}
		current = next
		retValues[i] = current
	}

	return retValues, nil
}
