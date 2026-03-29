package net

import (
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

func TestLookupSRV(t *testing.T) {
	srv, err := LookupSRV("_sip._udp.sip.voice.google.com")
	require.NoError(t, err)
	assert.Equal(t, uint16(5060), srv.Port)
}
