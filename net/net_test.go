package net

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func must(r interface{}, err error) interface{} {
	if err != nil {
		panic(err)
	}
	return r
}

func TestLookupIP(t *testing.T) {
	assert.Equal(t, "127.0.0.1", must(LookupIP("localhost")))
	assert.Equal(t, "198.41.0.4", must(LookupIP("a.root-servers.net")))
}

func TestLookupIPs(t *testing.T) {
	assert.Equal(t, []string{"127.0.0.1"}, must(LookupIPs("localhost")))
	assert.ElementsMatch(t, []string{"1.1.1.1", "1.0.0.1"}, must(LookupIPs("one.one.one.one")))
}

func TestLookupTXT(t *testing.T) {
	assert.NotEmpty(t, must(LookupTXT("example.com")))
}

func TestLookupCNAME(t *testing.T) {
	assert.Equal(t, "hairyhenderson.ca.", must(LookupCNAME("www.hairyhenderson.ca.")))
}

func TestLookupSRV(t *testing.T) {
	srv, err := LookupSRV("_sip._udp.sip.voice.google.com")
	assert.NoError(t, err)
	assert.Equal(t, uint16(5060), srv.Port)
}
