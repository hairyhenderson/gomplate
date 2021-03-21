package datasources

import (
	"context"
	"io/ioutil"
	"testing"

	"github.com/hairyhenderson/gomplate/v3/internal/config"
	"github.com/hairyhenderson/gomplate/v3/vault"

	"github.com/stretchr/testify/assert"
)

func TestVaultRequester(t *testing.T) {
	expected := "{\"value\":\"foo\"}\n"
	server, v := vault.MockServer(200, `{"data":`+expected+`}`)
	defer server.Close()

	r := &vaultRequester{}

	ctx := config.WithVaultClient(context.Background(), v)

	source := mustParseURL("vault:///secret/foo")

	resp, err := r.Request(ctx, source, nil)
	assert.NoError(t, err)
	defer resp.Body.Close()

	b, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, []byte(expected), b)

	source = mustParseURL("vault:///secret/foo/bar")
	resp, err = r.Request(ctx, source, nil)
	assert.NoError(t, err)
	defer resp.Body.Close()

	b, _ = ioutil.ReadAll(resp.Body)
	assert.Equal(t, []byte(expected), b)

	source = mustParseURL("vault:///secret/foo?param=value")
	resp, err = r.Request(ctx, source, nil)
	assert.NoError(t, err)
	defer resp.Body.Close()

	b, _ = ioutil.ReadAll(resp.Body)
	assert.Equal(t, []byte(expected), b)

	source = mustParseURL("vault:///secret/foo?param1=value1&param2=value2")
	resp, err = r.Request(ctx, source, nil)
	assert.NoError(t, err)
	defer resp.Body.Close()

	b, _ = ioutil.ReadAll(resp.Body)
	assert.Equal(t, []byte(expected), b)

	expected = "[\"one\",\"two\"]\n"
	server, v = vault.MockServer(200, `{"data":{"keys":`+expected+`}}`)
	defer server.Close()

	ctx = config.WithVaultClient(ctx, v)

	source = mustParseURL("vault:///secret/foo/")
	resp, err = r.Request(ctx, source, nil)
	assert.NoError(t, err)
	defer resp.Body.Close()

	b, _ = ioutil.ReadAll(resp.Body)
	assert.Equal(t, []byte(expected), b)
}
