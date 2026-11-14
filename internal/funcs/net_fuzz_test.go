package funcs

import (
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func FuzzCIDRSubnets(f *testing.F) {
	f.Add(2, "10.0.0.0/16")
	f.Add(8, "10.3.0.0/16")
	f.Add(1, "8000::/1")
	f.Add(-1, "10.0.0.0/16")
	f.Add(40, "10.0.0.0/16")
	f.Add(64, "::/0")
	f.Add(63, "::/0")

	f.Fuzz(func(t *testing.T, newbits int, prefix string) {
		network, err := netip.ParsePrefix(prefix)
		if err != nil {
			t.Skip()
		}

		// a valid request for a huge number of subnets only tests the
		// allocator, but invalid ones must fail before allocating anything
		fits := newbits >= 1 && network.Bits()+newbits <= network.Addr().BitLen()
		if fits && newbits > 12 && newbits <= 62 {
			t.Skip()
		}

		n := testNetNS(t)
		subnets, err := n.CIDRSubnets(newbits, network)
		if !fits || newbits > 62 {
			require.Error(t, err)
			return
		}

		require.NoError(t, err)
		require.Len(t, subnets, 1<<newbits)

		// every subnet is the right size, is inside the network, and comes
		// after the one before it
		for i, s := range subnets {
			assert.Equal(t, network.Bits()+newbits, s.Bits(), s)
			assert.True(t, network.Contains(s.Addr()), "%s isn't in %s", s, network)
			if i > 0 {
				assert.Negative(t, subnets[i-1].Addr().Compare(s.Addr()), "%s isn't after %s", s, subnets[i-1])
			}
		}
	})
}
