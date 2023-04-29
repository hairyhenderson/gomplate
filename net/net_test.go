package net

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func must(r interface{}, err error) interface{} {
	if err != nil {
		panic(err)
	}
	return r
}
func TestLookupIP(t *testing.T) {
	assert.Equal(t, "127.0.0.1", must(LookupIP("localhost")))
	assert.Equal(t, "93.184.216.34", must(LookupIP("example.com")))
}

func TestLookupIPs(t *testing.T) {
	assert.Equal(t, []string{"127.0.0.1"}, must(LookupIPs("localhost")))
	assert.Equal(t, []string{"93.184.216.34"}, must(LookupIPs("example.com")))
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
