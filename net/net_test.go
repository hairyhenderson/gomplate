package net

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLookupIP(t *testing.T) {
	assert.Equal(t, "127.0.0.1", LookupIP("localhost"))
	assert.Equal(t, "169.254.255.254", LookupIP("hostlocal.io"))
	assert.Equal(t, "93.184.216.34", LookupIP("example.com"))

}

func TestLookupIPs(t *testing.T) {
	assert.Equal(t, "127.0.0.1", LookupIPs("localhost")[0])
	assert.Equal(t, []string{"169.254.255.254"}, LookupIPs("hostlocal.io"))
	assert.Equal(t, []string{"93.184.216.34"}, LookupIPs("example.com"))
}

func TestLookupTXT(t *testing.T) {
	assert.NotEmpty(t, LookupTXT("example.com"))
}

func TestLookupCNAME(t *testing.T) {
	assert.Equal(t, "hairyhenderson.ca.", LookupCNAME("www.hairyhenderson.ca."))
}

func TestLookupSRV(t *testing.T) {
	assert.Equal(t, uint16(5060), LookupSRV("_sip._udp.sip.voice.google.com").Port)
}
