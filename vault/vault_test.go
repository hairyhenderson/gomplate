package vault

import (
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	os.Unsetenv("VAULT_ADDR")
	v, err := New(nil)
	assert.NoError(t, err)
	assert.Equal(t, "https://127.0.0.1:8200", v.client.Address())

	os.Setenv("VAULT_ADDR", "http://example.com:1234")
	defer os.Unsetenv("VAULT_ADDR")
	v, err = New(nil)
	assert.NoError(t, err)
	assert.Equal(t, "http://example.com:1234", v.client.Address())
	os.Unsetenv("VAULT_ADDR")

	u, _ := url.Parse("vault://vault.rocks:8200/secret/foo/bar")
	v, err = New(u)
	assert.NoError(t, err)
	assert.Equal(t, "https://vault.rocks:8200", v.client.Address())

	u, _ = url.Parse("vault+https://vault.rocks:8200/secret/foo/bar")
	v, err = New(u)
	assert.NoError(t, err)
	assert.Equal(t, "https://vault.rocks:8200", v.client.Address())

	u, _ = url.Parse("vault+http://vault.rocks:8200/secret/foo/bar")
	v, err = New(u)
	assert.NoError(t, err)
	assert.Equal(t, "http://vault.rocks:8200", v.client.Address())
}

func TestRead(t *testing.T) {
	server, v := MockServer(404, "")
	defer server.Close()
	val, err := v.Read("secret/bogus")
	assert.Empty(t, val)
	assert.NoError(t, err)

	expected := "{\"value\":\"foo\"}\n"
	server, v = MockServer(200, `{"data":`+expected+`}`)
	defer server.Close()
	val, err = v.Read("s")
	assert.Equal(t, expected, string(val))
	assert.NoError(t, err)
}

func TestWrite(t *testing.T) {
	server, v := MockServer(404, "Not Found")
	defer server.Close()
	val, err := v.Write("secret/bogus", nil)
	assert.Empty(t, val)
	assert.Error(t, err)

	expected := "{\"value\":\"foo\"}\n"
	server, v = MockServer(200, `{"data":`+expected+`}`)
	defer server.Close()
	val, err = v.Write("s", nil)
	assert.Equal(t, expected, string(val))
	assert.NoError(t, err)
}
