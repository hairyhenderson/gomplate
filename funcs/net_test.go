package funcs

import (
	"context"
	"strconv"
	"testing"

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
