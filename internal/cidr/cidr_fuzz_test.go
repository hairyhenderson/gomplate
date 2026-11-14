package cidr

import (
	"math"
	"math/big"
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fuzzNum builds a big.Int of any size from the fuzzer's bytes.
func fuzzNum(b []byte, neg bool) *big.Int {
	n := new(big.Int).SetBytes(b)
	if neg {
		n.Neg(n)
	}
	return n
}

func FuzzHostBig(f *testing.F) {
	f.Add("10.12.112.0/20", []byte{16}, false)
	f.Add("10.12.112.0/20", []byte{1}, true)
	f.Add("10.12.112.0/20", []byte{0x10, 0x00}, false)
	f.Add("fd00:fd12:3456:7890::/56", []byte{0x22}, false)
	f.Add("10.0.0.0/8", []byte{1, 0, 0, 0, 0, 0, 0, 0, 0}, false)

	f.Fuzz(func(t *testing.T, prefix string, b []byte, neg bool) {
		base, err := netip.ParsePrefix(prefix)
		if err != nil || len(b) > 32 {
			t.Skip()
		}

		num := fuzzNum(b, neg)
		orig := new(big.Int).Set(num)

		// the number of hosts in the network
		size := new(big.Int).Lsh(big.NewInt(1), uint(base.Addr().BitLen()-base.Bits()))
		// negative numbers count back from the end of the network
		want := new(big.Int).Set(num)
		if want.Sign() < 0 {
			want.Add(want, size)
		}
		fits := want.Sign() >= 0 && want.Cmp(size) < 0

		ip, err := HostBig(base, num)
		assert.Zero(t, orig.Cmp(num), "HostBig modified num from %s to %s", orig, num)
		if !fits {
			require.Error(t, err)
			return
		}

		require.NoError(t, err)
		assert.True(t, base.Contains(ip), "%s isn't in %s", ip, base)

		// the host's offset from the start of the network is the number
		start, _ := ipToInt(base.Masked().Addr())
		got, _ := ipToInt(ip)
		got.Sub(got, start)
		assert.Zero(t, want.Cmp(got), "host %s is %s from the start of %s, not %s", ip, got, base, want)
	})
}

func FuzzSubnetBig(f *testing.F) {
	f.Add("10.3.0.0/16", 8, []byte{5}, false)
	f.Add("10.3.0.0/16", 16, []byte{0xff, 0xff}, false)
	f.Add("10.3.0.0/16", 17, []byte{0}, false)
	f.Add("fd00::/56", 8, []byte{0x22}, false)
	f.Add("::/0", 64, []byte{1, 0, 0, 0, 0, 0, 0, 0, 0}, false)
	f.Add("10.3.0.0/16", 8, []byte{1}, true)
	f.Add("10.3.0.0/16", -8, []byte{0}, false)
	f.Add("10.3.0.0/16", math.MaxInt, []byte{0}, false)

	f.Fuzz(func(t *testing.T, prefix string, newBits int, b []byte, neg bool) {
		base, err := netip.ParsePrefix(prefix)
		if err != nil || len(b) > 32 {
			t.Skip()
		}

		num := fuzzNum(b, neg)
		orig := new(big.Int).Set(num)

		fits := newBits >= 0 && newBits <= base.Addr().BitLen()-base.Bits() &&
			num.Sign() >= 0 && num.BitLen() <= newBits

		subnet, err := SubnetBig(base, newBits, num)
		assert.Zero(t, orig.Cmp(num), "SubnetBig modified num from %s to %s", orig, num)
		if !fits {
			require.Error(t, err)
			return
		}

		require.NoError(t, err)
		assert.Equal(t, base.Bits()+newBits, subnet.Bits())
		assert.Equal(t, subnet.Masked(), subnet, "%s isn't a network address", subnet)
		assert.True(t, base.Contains(subnet.Addr()), "%s isn't in %s", subnet, base)

		// the subnet's offset from the start of base, in subnets, is the number
		start, _ := ipToInt(base.Masked().Addr())
		got, _ := ipToInt(subnet.Addr())
		got.Sub(got, start)
		got.Rsh(got, uint(subnet.Addr().BitLen()-subnet.Bits()))
		assert.Zero(t, orig.Cmp(got), "%s is subnet %s of %s, not %s", subnet, got, base, orig)
	})
}
