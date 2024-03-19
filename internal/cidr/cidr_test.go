package cidr

import (
	"fmt"
	"math/big"
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubnetBig(t *testing.T) {
	cases := []struct {
		base string
		num  *big.Int
		out  string
		bits int
		err  bool
	}{
		{
			base: "192.168.2.0/20",
			bits: 4,
			num:  big.NewInt(int64(6)),
			out:  "192.168.6.0/24",
		},
		{
			base: "192.168.2.0/20",
			bits: 4,
			num:  big.NewInt(int64(0)),
			out:  "192.168.0.0/24",
		},
		{
			base: "192.168.0.0/31",
			bits: 1,
			num:  big.NewInt(int64(1)),
			out:  "192.168.0.1/32",
		},
		{
			base: "192.168.0.0/21",
			bits: 4,
			num:  big.NewInt(int64(7)),
			out:  "192.168.3.128/25",
		},
		{
			base: "fe80::/48",
			bits: 16,
			num:  big.NewInt(int64(6)),
			out:  "fe80:0:0:6::/64",
		},
		{
			base: "fe80::/48",
			bits: 33,
			num:  big.NewInt(int64(6)),
			out:  "fe80::3:0:0:0/81",
		},
		{
			base: "fe80::/49",
			bits: 16,
			num:  big.NewInt(int64(7)),
			out:  "fe80:0:0:3:8000::/65",
		},
		{
			base: "192.168.2.0/31",
			bits: 2,
			num:  big.NewInt(int64(0)),
			err:  true, // not enough bits to expand into
		},
		{
			base: "fe80::/126",
			bits: 4,
			num:  big.NewInt(int64(0)),
			err:  true, // not enough bits to expand into
		},
		{
			base: "192.168.2.0/24",
			bits: 4,
			num:  big.NewInt(int64(16)),
			err:  true, // can't fit 16 into 4 bits
		},
	}

	for _, testCase := range cases {
		t.Run(fmt.Sprintf("SubnetBig(%#v,%#v,%#v)", testCase.base, testCase.bits, testCase.num), func(t *testing.T) {
			base := netip.MustParsePrefix(testCase.base)

			subnet, err := SubnetBig(base, testCase.bits, testCase.num)
			if testCase.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.out, subnet.String())
			}
		})
	}
}

func TestHostBig(t *testing.T) {
	cases := []struct {
		prefix string
		num    *big.Int
		out    string
		err    bool
	}{
		{
			prefix: "192.168.2.0/20",
			num:    big.NewInt(int64(6)),
			out:    "192.168.0.6",
		},
		{
			prefix: "192.168.0.0/20",
			num:    big.NewInt(int64(257)),
			out:    "192.168.1.1",
		},
		{
			prefix: "2001:db8::/32",
			num:    big.NewInt(int64(1)),
			out:    "2001:db8::1",
		},
		{
			prefix: "192.168.1.0/24",
			num:    big.NewInt(int64(256)),
			err:    true, // only 0-255 will fit in 8 bits
		},
		{
			prefix: "192.168.0.0/30",
			num:    big.NewInt(int64(-3)),
			out:    "192.168.0.1", // 4 address (0-3) in 2 bits; 3rd from end = 1
		},
		{
			prefix: "192.168.0.0/30",
			num:    big.NewInt(int64(-4)),
			out:    "192.168.0.0", // 4 address (0-3) in 2 bits; 4th from end = 0
		},
		{
			prefix: "192.168.0.0/30",
			num:    big.NewInt(int64(-5)),
			err:    true, // 4 address (0-3) in 2 bits; cannot accommodate 5
		},
		{
			prefix: "fd9d:bc11:4020::/64",
			num:    big.NewInt(int64(2)),
			out:    "fd9d:bc11:4020::2",
		},
		{
			prefix: "fd9d:bc11:4020::/64",
			num:    big.NewInt(int64(-2)),
			out:    "fd9d:bc11:4020:0:ffff:ffff:ffff:fffe",
		},
	}

	for _, testCase := range cases {
		t.Run(fmt.Sprintf("HostBig(%v,%v)", testCase.prefix, testCase.num), func(t *testing.T) {
			network := netip.MustParsePrefix(testCase.prefix)

			gotIP, err := HostBig(network, testCase.num)
			if testCase.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.out, gotIP.String())
			}
		})
	}
}

func TestPreviousNextSubnet(t *testing.T) {
	testCases := []struct {
		next, prev string
		overflow   bool
	}{
		{"10.0.0.0/24", "9.255.255.0/24", false},
		{"100.0.0.0/26", "99.255.255.192/26", false},
		{"0.0.0.0/26", "255.255.255.192/26", true},
		{"2001:db8:e000::/36", "2001:db8:d000::/36", false},
		{"::/64", "ffff:ffff:ffff:ffff::/64", true},
	}
	for _, tc := range testCases {
		c1 := netip.MustParsePrefix(tc.next)
		c2 := netip.MustParsePrefix(tc.prev)
		mask := c1.Bits()

		p1, rollback := PreviousSubnet(c1, mask)
		if tc.overflow {
			assert.True(t, rollback)
			continue
		}

		assert.Equal(t, c2.String(), p1.String())
	}

	for _, tc := range testCases {
		c1 := netip.MustParsePrefix(tc.next)
		c2 := netip.MustParsePrefix(tc.prev)
		mask := c1.Bits()

		n1, rollover := NextSubnet(c2, mask)
		if tc.overflow {
			assert.True(t, rollover)
			continue
		}
		assert.Equal(t, c1, n1)
	}
}
