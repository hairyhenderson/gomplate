package funcs

import (
	"context"
	stdnet "net"
	"net/netip"
	"strconv"
	"testing"

	"github.com/hairyhenderson/gomplate/v3/internal/config"
	"github.com/stretchr/testify/assert"
	"inet.af/netaddr"
)

func TestCreateNetFuncs(t *testing.T) {
	for i := 0; i < 10; i++ {
		// Run this a bunch to catch race conditions
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			fmap := CreateNetFuncs(ctx)
			actual := fmap["net"].(func() interface{})

			assert.Same(t, ctx, actual().(*NetFuncs).ctx)
		})
	}
}

func TestNetLookupIP(t *testing.T) {
	n := NetFuncs{}
	assert.Equal(t, "127.0.0.1", must(n.LookupIP("localhost")))
}

func TestParseIP(t *testing.T) {
	n := NetFuncs{}
	_, err := n.ParseIP("not an IP")
	assert.Error(t, err)

	ip, err := n.ParseIP("2001:470:20::2")
	assert.NoError(t, err)
	assert.Equal(t, netaddr.IPFrom16([16]byte{
		0x20, 0x01, 0x04, 0x70,
		0, 0x20, 0, 0,
		0, 0, 0, 0,
		0, 0, 0, 0x02,
	}), ip)
}

func TestParseIPPrefix(t *testing.T) {
	n := NetFuncs{}
	_, err := n.ParseIPPrefix("not an IP")
	assert.Error(t, err)

	_, err = n.ParseIPPrefix("1.1.1.1")
	assert.Error(t, err)

	ipprefix, err := n.ParseIPPrefix("192.168.0.2/28")
	assert.NoError(t, err)
	assert.Equal(t, "192.168.0.0/28", ipprefix.Masked().String())
}

func TestParseIPRange(t *testing.T) {
	n := NetFuncs{}
	_, err := n.ParseIPRange("not an IP")
	assert.Error(t, err)

	_, err = n.ParseIPRange("1.1.1.1")
	assert.Error(t, err)

	iprange, err := n.ParseIPRange("192.168.0.2-192.168.23.255")
	assert.NoError(t, err)
	assert.Equal(t, "192.168.0.2-192.168.23.255", iprange.String())
}

func testNetNS() *NetFuncs {
	ctx := context.Background()
	cfg := config.FromContext(ctx)
	cfg.Experimental = true
	ctx = config.ContextWithConfig(ctx, cfg)

	return &NetFuncs{
		ctx: ctx,
	}
}

func TestCIDRHost(t *testing.T) {
	n := testNetNS()

	// net.IPNet
	_, netIP, _ := stdnet.ParseCIDR("10.12.127.0/20")

	ip, err := n.CIDRHost(16, netIP)
	assert.NoError(t, err)
	assert.Equal(t, "10.12.112.16", ip.String())

	ip, err = n.CIDRHost(268, netIP)
	assert.NoError(t, err)
	assert.Equal(t, "10.12.113.12", ip.String())

	_, netIP, _ = stdnet.ParseCIDR("fd00:fd12:3456:7890:00a2::/72")
	ip, err = n.CIDRHost(34, netIP)
	assert.NoError(t, err)
	assert.Equal(t, "fd00:fd12:3456:7890::22", ip.String())

	// inet.af/netaddr.IPPrefix
	ipPrefix, _ := n.ParseIPPrefix("10.12.127.0/20")

	ip, err = n.CIDRHost(16, ipPrefix)
	assert.NoError(t, err)
	assert.Equal(t, "10.12.112.16", ip.String())

	ip, err = n.CIDRHost(268, ipPrefix)
	assert.NoError(t, err)
	assert.Equal(t, "10.12.113.12", ip.String())

	ipPrefix, _ = n.ParseIPPrefix("fd00:fd12:3456:7890:00a2::/72")
	ip, err = n.CIDRHost(34, ipPrefix)
	assert.NoError(t, err)
	assert.Equal(t, "fd00:fd12:3456:7890::22", ip.String())

	// net/netip.Prefix
	prefix := netip.MustParsePrefix("10.12.127.0/20")

	ip, err = n.CIDRHost(16, prefix)
	assert.NoError(t, err)
	assert.Equal(t, "10.12.112.16", ip.String())

	ip, err = n.CIDRHost(268, prefix)
	assert.NoError(t, err)
	assert.Equal(t, "10.12.113.12", ip.String())

	prefix = netip.MustParsePrefix("fd00:fd12:3456:7890:00a2::/72")
	ip, err = n.CIDRHost(34, prefix)
	assert.NoError(t, err)
	assert.Equal(t, "fd00:fd12:3456:7890::22", ip.String())
}

func TestCIDRNetmask(t *testing.T) {
	n := testNetNS()

	ip, err := n.CIDRNetmask("10.0.0.0/12")
	assert.NoError(t, err)
	assert.Equal(t, "255.240.0.0", ip.String())

	ip, err = n.CIDRNetmask("fd00:fd12:3456:7890:00a2::/72")
	assert.NoError(t, err)
	assert.Equal(t, "ffff:ffff:ffff:ffff:ff00::", ip.String())
}

func TestCIDRSubnets(t *testing.T) {
	n := testNetNS()
	network := netip.MustParsePrefix("10.0.0.0/16")

	subnets, err := n.CIDRSubnets(-1, network)
	assert.Nil(t, subnets)
	assert.Error(t, err)

	subnets, err = n.CIDRSubnets(2, network)
	assert.NoError(t, err)
	assert.Len(t, subnets, 4)
	assert.Equal(t, "10.0.0.0/18", subnets[0].String())
	assert.Equal(t, "10.0.64.0/18", subnets[1].String())
	assert.Equal(t, "10.0.128.0/18", subnets[2].String())
	assert.Equal(t, "10.0.192.0/18", subnets[3].String())
}

func TestCIDRSubnetSizes(t *testing.T) {
	n := testNetNS()
	network := netip.MustParsePrefix("10.1.0.0/16")

	subnets, err := n.CIDRSubnetSizes(network)
	assert.Nil(t, subnets)
	assert.Error(t, err)

	subnets, err = n.CIDRSubnetSizes(32, network)
	assert.Nil(t, subnets)
	assert.Error(t, err)

	subnets, err = n.CIDRSubnetSizes(-1, network)
	assert.Nil(t, subnets)
	assert.Error(t, err)

	subnets, err = n.CIDRSubnetSizes(4, 4, 8, 4, network)
	assert.NoError(t, err)
	assert.Len(t, subnets, 4)
	assert.Equal(t, "10.1.0.0/20", subnets[0].String())
	assert.Equal(t, "10.1.16.0/20", subnets[1].String())
	assert.Equal(t, "10.1.32.0/24", subnets[2].String())
	assert.Equal(t, "10.1.48.0/20", subnets[3].String())
}
