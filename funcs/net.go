package funcs

import (
	"context"
	"math/big"
	stdnet "net"
	"net/netip"

	"github.com/apparentlymart/go-cidr/cidr"
	"github.com/flanksource/gomplate/v3/conv"
	"github.com/flanksource/gomplate/v3/net"
	"github.com/pkg/errors"
)

// NetNS - the net namespace
// Deprecated: don't use
func NetNS() *NetFuncs {
	return &NetFuncs{}
}

// AddNetFuncs -
// Deprecated: use CreateNetFuncs instead
func AddNetFuncs(f map[string]interface{}) {
	for k, v := range CreateNetFuncs(context.Background()) {
		f[k] = v
	}
}

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
func (f NetFuncs) ParseIP(ip interface{}) (netip.Addr, error) {
	return netip.ParseAddr(conv.ToString(ip))
}

// ParseIPPrefix -
func (f NetFuncs) ParseIPPrefix(ipprefix interface{}) (netip.Prefix, error) {
	return netip.ParsePrefix(conv.ToString(ipprefix))
}

// StdParseIP -
func (f NetFuncs) StdParseIP(prefix interface{}) (stdnet.IP, error) {
	ip := stdnet.ParseIP(conv.ToString(prefix))
	if ip == nil {
		return nil, errors.Errorf("invalid IP address")
	}
	return ip, nil
}

func (f NetFuncs) stdParseCIDR(prefix interface{}) (*stdnet.IPNet, error) {
	if n, ok := prefix.(*stdnet.IPNet); ok {
		return n, nil
	}

	_, network, err := stdnet.ParseCIDR(conv.ToString(prefix))
	return network, err
}

// StdParseCIDR -
func (f NetFuncs) StdParseCIDR(prefix interface{}) (*stdnet.IPNet, error) {
	return f.stdParseCIDR(prefix)
}

func (f NetFuncs) CIDRHost(hostnum interface{}, prefix interface{}) (*stdnet.IP, error) {
	return f.CidrHost(hostnum, prefix)
}

// CidrHost -
func (f NetFuncs) CidrHost(hostnum interface{}, prefix interface{}) (*stdnet.IP, error) {
	network, err := f.stdParseCIDR(prefix)
	if err != nil {
		return nil, err
	}

	ip, err := cidr.HostBig(network, big.NewInt(conv.ToInt64(hostnum)))
	return &ip, err
}

// CidrNetmask -
func (f NetFuncs) CIDRNetmask(prefix interface{}) (*stdnet.IP, error) {
	return f.CidrNetmask(prefix)
}

// CidrNetmask -
func (f NetFuncs) CidrNetmask(prefix interface{}) (*stdnet.IP, error) {
	network, err := f.stdParseCIDR(prefix)
	if err != nil {
		return nil, err
	}

	if len(network.IP) != stdnet.IPv4len {
		return nil, errors.Errorf("only IPv4 networks are supported")
	}

	netmask := stdnet.IP(network.Mask)
	return &netmask, nil
}

// CidrSubnets -
func (f NetFuncs) CIDRSubnets(newbits interface{}, prefix interface{}) ([]*stdnet.IPNet, error) {
	return f.CidrSubnets(newbits, prefix)
}

// CidrSubnets -
func (f NetFuncs) CidrSubnets(newbits interface{}, prefix interface{}) ([]*stdnet.IPNet, error) {
	network, err := f.stdParseCIDR(prefix)
	if err != nil {
		return nil, err
	}

	nBits := conv.ToInt(newbits)
	if nBits < 1 {
		return nil, errors.Errorf("must extend prefix by at least one bit")
	}

	maxNetNum := int64(1 << uint64(nBits))
	retValues := make([]*stdnet.IPNet, maxNetNum)
	for i := int64(0); i < maxNetNum; i++ {
		subnet, err := cidr.SubnetBig(network, nBits, big.NewInt(i))
		if err != nil {
			return nil, err
		}
		retValues[i] = subnet
	}

	return retValues, nil
}

// CidrSubnetSizes -
func (f NetFuncs) CIDRSubnetSizes(args ...interface{}) ([]*stdnet.IPNet, error) {
	return f.CidrSubnetSizes(args...)
}

// CidrSubnetSizes -
func (f NetFuncs) CidrSubnetSizes(args ...interface{}) ([]*stdnet.IPNet, error) {
	if len(args) < 2 {
		return nil, errors.Errorf("wrong number of args: want 2 or more, got %d", len(args))
	}

	network, err := f.stdParseCIDR(args[len(args)-1])
	if err != nil {
		return nil, err
	}
	newbits := conv.ToInts(args[:len(args)-1]...)

	startPrefixLen, _ := network.Mask.Size()
	firstLength := newbits[0]

	firstLength += startPrefixLen
	retValues := make([]*stdnet.IPNet, len(newbits))

	current, _ := cidr.PreviousSubnet(network, firstLength)

	for i, length := range newbits {
		if length < 1 {
			return nil, errors.Errorf("must extend prefix by at least one bit")
		}
		// For portability with 32-bit systems where the subnet number
		// will be a 32-bit int, we only allow extension of 32 bits in
		// one call even if we're running on a 64-bit machine.
		// (Of course, this is significant only for IPv6.)
		if length > 32 {
			return nil, errors.Errorf("may not extend prefix by more than 32 bits")
		}

		length += startPrefixLen
		if length > (len(network.IP) * 8) {
			protocol := "IP"
			switch len(network.IP) {
			case stdnet.IPv4len:
				protocol = "IPv4"
			case stdnet.IPv6len:
				protocol = "IPv6"
			}
			return nil, errors.Errorf("would extend prefix to %d bits, which is too long for an %s address", length, protocol)
		}

		next, rollover := cidr.NextSubnet(current, length)
		if rollover || !network.Contains(next.IP) {
			// If we run out of suffix bits in the base CIDR prefix then
			// NextSubnet will start incrementing the prefix bits, which
			// we don't allow because it would then allocate addresses
			// outside of the caller's given prefix.
			return nil, errors.Errorf("not enough remaining address space for a subnet with a prefix of %d bits after %s", length, current.String())
		}
		current = next
		retValues[i] = current
	}

	return retValues, nil
}
