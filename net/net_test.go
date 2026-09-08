package net

import (
	"net"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func must(r any, err error) any {
	if err != nil {
		panic(err)
	}
	return r
}

func TestLookupIP(t *testing.T) {
	assert.Equal(t, "127.0.0.1", must(LookupIP(t.Context(), "localhost")))
}

func TestLookupIPs(t *testing.T) {
	assert.Equal(t, []string{"127.0.0.1"}, must(LookupIPs(t.Context(), "localhost")))
}

func BenchmarkLookupIPs(b *testing.B) {
	for b.Loop() {
		must(LookupIPs(b.Context(), "localhost"))
	}
}

func TestLookupTXT(t *testing.T) {
	assert.NotEmpty(t, must(LookupTXT("example.com")))
}

func TestLookupCNAME(t *testing.T) {
	assert.Equal(t, "hairyhenderson.ca.", must(LookupCNAME("www.hairyhenderson.ca.")))
}

func TestGenerateMAC(t *testing.T) {
	// no prefix: a valid, locally administered unicast address
	first, err := GenerateMAC("", nil)
	require.NoError(t, err)
	hw, err := net.ParseMAC(first)
	require.NoError(t, err)
	require.Len(t, hw, macLen)
	assert.Equal(t, byte(0x02), hw[0]&0x02, "local bit should be set")
	assert.Equal(t, byte(0x00), hw[0]&0x01, "multicast bit should be clear")

	// without a seed each call differs
	second, err := GenerateMAC("", nil)
	require.NoError(t, err)
	assert.NotEqual(t, first, second)

	// a prefix is respected, and separators don't matter
	for _, p := range []string{"aa:bb:cc", "aa-bb-cc", "aabbcc"} {
		mac, perr := GenerateMAC(p, nil)
		require.NoError(t, perr)
		assert.True(t, strings.HasPrefix(mac, "aa:bb:cc:"), "prefix %q gave %q", p, mac)
	}

	// a seed makes the output deterministic for a given prefix
	seeded, err := GenerateMAC("aa:bb", []byte("gomplate"))
	require.NoError(t, err)
	again, err := GenerateMAC("aa:bb", []byte("gomplate"))
	require.NoError(t, err)
	assert.Equal(t, seeded, again)

	other, err := GenerateMAC("aa:bb", []byte("different"))
	require.NoError(t, err)
	assert.NotEqual(t, seeded, other)

	// a full 6-octet prefix leaves nothing to randomize
	full, err := GenerateMAC("01:02:03:04:05:06", nil)
	require.NoError(t, err)
	assert.Equal(t, "01:02:03:04:05:06", full)

	// invalid inputs
	_, err = GenerateMAC("not hex", nil)
	require.Error(t, err)
	_, err = GenerateMAC("aa:b", nil) // odd number of hex digits
	require.Error(t, err)
	_, err = GenerateMAC("aa:bb:cc:dd:ee:ff:00", nil) // too long
	require.Error(t, err)
}

func TestLookupSRV(t *testing.T) {
	srv, err := LookupSRV("_sip._udp.sip.voice.google.com")
	require.NoError(t, err)
	assert.Equal(t, uint16(5060), srv.Port)
}
