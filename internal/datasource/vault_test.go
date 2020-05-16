package datasource

import (
	"context"
	"net/url"
	"testing"

	"github.com/hairyhenderson/gomplate/v3/vault"
	"github.com/stretchr/testify/assert"
)

func TestReadVault(t *testing.T) {
	expected := []byte("{\"value\":\"foo\"}\n")
	server, vc := vault.MockServer(200, `{"data":`+string(expected)+`}`)
	defer server.Close()

	u := &url.URL{Scheme: "vault", Path: "/secret/foo"}
	ctx := context.Background()
	v := &Vault{vc}

	r, err := v.Read(ctx, u)
	assert.NoError(t, err)
	assert.Equal(t, expected, r.Bytes)
	assert.Equal(t, jsonMimetype, r.MediaType)

	r, err = v.Read(ctx, u, "bar")
	assert.NoError(t, err)
	assert.Equal(t, expected, r.Bytes)
	assert.Equal(t, jsonMimetype, r.MediaType)

	r, err = v.Read(ctx, u, "?param=value")
	assert.NoError(t, err)
	assert.Equal(t, expected, r.Bytes)
	assert.Equal(t, jsonMimetype, r.MediaType)

	u, _ = url.Parse("vault:///secret/foo?param1=value1&param2=value2")
	r, err = v.Read(ctx, u)
	assert.NoError(t, err)
	assert.Equal(t, expected, r.Bytes)
	assert.Equal(t, jsonMimetype, r.MediaType)

	expected = []byte("[\"one\",\"two\"]\n")
	server, v.vc = vault.MockServer(200, `{"data":{"keys":`+string(expected)+`}}`)
	defer server.Close()

	u, _ = url.Parse("vault:///secret/foo/")
	r, err = v.Read(ctx, u)
	assert.NoError(t, err)
	assert.Equal(t, expected, r.Bytes)
	assert.Equal(t, jsonArrayMimetype, r.MediaType)
}
