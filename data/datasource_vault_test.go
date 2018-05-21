package data

import (
	"net/url"
	"testing"

	"github.com/hairyhenderson/gomplate/vault"
	"github.com/stretchr/testify/assert"
)

func TestReadVault(t *testing.T) {
	expected := "{\"value\":\"foo\"}\n"
	server, v := vault.MockServer(200, `{"data":`+expected+`}`)
	defer server.Close()

	source := &Source{
		Alias: "foo",
		URL:   &url.URL{Scheme: "vault", Path: "/secret/foo"},
		Ext:   "",
		Type:  "text/plain",
		VC:    v,
	}

	r, err := readVault(source)
	assert.NoError(t, err)
	assert.Equal(t, []byte(expected), r)

	r, err = readVault(source, "bar")
	assert.NoError(t, err)
	assert.Equal(t, []byte(expected), r)

	r, err = readVault(source, "?param=value")
	assert.NoError(t, err)
	assert.Equal(t, []byte(expected), r)

	source.URL, _ = url.Parse("vault:///secret/foo?param1=value1&param2=value2")
	r, err = readVault(source)
	assert.NoError(t, err)
	assert.Equal(t, []byte(expected), r)

	expected = "[\"one\",\"two\"]\n"
	server, source.VC = vault.MockServer(200, `{"data":{"keys":`+expected+`}}`)
	defer server.Close()
	source.URL, _ = url.Parse("vault:///secret/foo/")
	r, err = readVault(source)
	assert.NoError(t, err)
	assert.Equal(t, []byte(expected), r)
}
