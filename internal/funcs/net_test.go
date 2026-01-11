package funcs

import (
	"context"
	stdnet "net"
	"net/netip"
	"strconv"
	"testing"

	"github.com/hairyhenderson/gomplate/v4/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateNetFuncs(t *testing.T) {
	t.Parallel()

	for i := range 10 {
		// Run this a bunch to catch race conditions
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			fmap := CreateNetFuncs(ctx)
			actual := fmap["net"].(func() any)

			assert.Equal(t, ctx, actual().(*NetFuncs).ctx)
		})
	}
}

func TestNetLookupIP(t *testing.T) {
	t.Parallel()

	n := NetFuncs{}
	assert.Equal(t, "127.0.0.1", must(n.LookupIP("localhost")))
}

func TestParseAddr(t *testing.T) {
	t.Parallel()

	n := testNetNS()
	_, err := n.ParseAddr("not an IP")
	require.Error(t, err)

	ip, err := n.ParseAddr("2001:470:20::2")
	require.NoError(t, err)
	assert.Equal(t, netip.AddrFrom16([16]byte{
		0x20, 0x01, 0x04, 0x70,
		0, 0x20, 0, 0,
		0, 0, 0, 0,
		0, 0, 0, 0x02,
	}), ip)
}

func TestParsePrefix(t *testing.T) {
	t.Parallel()

	n := testNetNS()
	_, err := n.ParsePrefix("not an IP")
	require.Error(t, err)

	_, err = n.ParsePrefix("1.1.1.1")
	require.Error(t, err)

	ipprefix, err := n.ParsePrefix("192.168.0.2/28")
	require.NoError(t, err)
	assert.Equal(t, "192.168.0.0/28", ipprefix.Masked().String())
}

func TestParseRange(t *testing.T) {
	t.Parallel()

	n := testNetNS()
	_, err := n.ParseRange("not an IP")
	require.Error(t, err)

	_, err = n.ParseRange("1.1.1.1")
	require.Error(t, err)

	iprange, err := n.ParseRange("192.168.0.2-192.168.23.255")
	require.NoError(t, err)
	assert.Equal(t, "192.168.0.2-192.168.23.255", iprange.String())
}

func testNetNS() *NetFuncs {
	return &NetFuncs{ctx: config.SetExperimental(context.Background())}
}

func TestCIDRHost(t *testing.T) {
	n := testNetNS()

	// net.IPNet
	_, netIP, _ := stdnet.ParseCIDR("10.12.127.0/20")

	ip, err := n.CIDRHost(16, netIP)
	require.NoError(t, err)
	assert.Equal(t, "10.12.112.16", ip.String())

	ip, err = n.CIDRHost(268, netIP)
	require.NoError(t, err)
	assert.Equal(t, "10.12.113.12", ip.String())

	_, netIP, _ = stdnet.ParseCIDR("fd00:fd12:3456:7890:00a2::/72")
	ip, err = n.CIDRHost(34, netIP)
	require.NoError(t, err)
	assert.Equal(t, "fd00:fd12:3456:7890::22", ip.String())

	// net/netip.Prefix
	prefix := netip.MustParsePrefix("10.12.127.0/20")

	ip, err = n.CIDRHost(16, prefix)
	require.NoError(t, err)
	assert.Equal(t, "10.12.112.16", ip.String())

	ip, err = n.CIDRHost(268, prefix)
	require.NoError(t, err)
	assert.Equal(t, "10.12.113.12", ip.String())

	prefix = netip.MustParsePrefix("fd00:fd12:3456:7890:00a2::/72")
	ip, err = n.CIDRHost(34, prefix)
	require.NoError(t, err)
	assert.Equal(t, "fd00:fd12:3456:7890::22", ip.String())
}

func TestCIDRNetmask(t *testing.T) {
	n := testNetNS()

	ip, err := n.CIDRNetmask("10.0.0.0/12")
	require.NoError(t, err)
	assert.Equal(t, "255.240.0.0", ip.String())

	ip, err = n.CIDRNetmask("fd00:fd12:3456:7890:00a2::/72")
	require.NoError(t, err)
	assert.Equal(t, "ffff:ffff:ffff:ffff:ff00::", ip.String())
}

func TestCIDRSubnets(t *testing.T) {
	n := testNetNS()
	network := netip.MustParsePrefix("10.0.0.0/16")

	subnets, err := n.CIDRSubnets(-1, network)
	require.Error(t, err)
	assert.Nil(t, subnets)

	subnets, err = n.CIDRSubnets(2, network)
	require.NoError(t, err)
	assert.Len(t, subnets, 4)
	assert.Equal(t, "10.0.0.0/18", subnets[0].String())
	assert.Equal(t, "10.0.64.0/18", subnets[1].String())
	assert.Equal(t, "10.0.128.0/18", subnets[2].String())
	assert.Equal(t, "10.0.192.0/18", subnets[3].String())
}

func TestCIDRSubnetSizes(t *testing.T) {
	n := testNetNS()

	subnets, err := n.CIDRSubnetSizes(netip.MustParsePrefix("10.1.0.0/16"))
	require.Error(t, err)
	assert.Nil(t, subnets)

	subnets, err = n.CIDRSubnetSizes(32, netip.MustParsePrefix("10.1.0.0/16"))
	require.Error(t, err)
	assert.Nil(t, subnets)

	subnets, err = n.CIDRSubnetSizes(127, netip.MustParsePrefix("ffff::/48"))
	require.Error(t, err)
	assert.Nil(t, subnets)

	subnets, err = n.CIDRSubnetSizes(-1, netip.MustParsePrefix("10.1.0.0/16"))
	require.Error(t, err)
	assert.Nil(t, subnets)

	network := netip.MustParsePrefix("8000::/1")
	subnets, err = n.CIDRSubnetSizes(1, 2, 2, network)
	require.NoError(t, err)
	assert.Len(t, subnets, 3)
	assert.Equal(t, "8000::/2", subnets[0].String())
	assert.Equal(t, "c000::/3", subnets[1].String())
	assert.Equal(t, "e000::/3", subnets[2].String())

	network = netip.MustParsePrefix("10.1.0.0/16")
	subnets, err = n.CIDRSubnetSizes(4, 4, 8, 4, network)
	require.NoError(t, err)
	assert.Len(t, subnets, 4)
	assert.Equal(t, "10.1.0.0/20", subnets[0].String())
	assert.Equal(t, "10.1.16.0/20", subnets[1].String())
	assert.Equal(t, "10.1.32.0/24", subnets[2].String())
	assert.Equal(t, "10.1.48.0/20", subnets[3].String())

	network = netip.MustParsePrefix("2016:1234:5678:9abc:ffff:ffff:ffff:cafe/64")
	subnets, err = n.CIDRSubnetSizes(2, 2, 3, 3, 6, 6, 8, 10, network)
	require.NoError(t, err)
	assert.Len(t, subnets, 8)
	assert.Equal(t, "2016:1234:5678:9abc::/66", subnets[0].String())
	assert.Equal(t, "2016:1234:5678:9abc:4000::/66", subnets[1].String())
	assert.Equal(t, "2016:1234:5678:9abc:8000::/67", subnets[2].String())
	assert.Equal(t, "2016:1234:5678:9abc:a000::/67", subnets[3].String())
	assert.Equal(t, "2016:1234:5678:9abc:c000::/70", subnets[4].String())
	assert.Equal(t, "2016:1234:5678:9abc:c400::/70", subnets[5].String())
	assert.Equal(t, "2016:1234:5678:9abc:c800::/72", subnets[6].String())
	assert.Equal(t, "2016:1234:5678:9abc:c900::/74", subnets[7].String())
}
