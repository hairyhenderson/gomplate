package vault

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogin(t *testing.T) {
	server, v := MockServer(404, "Not Found")
	defer server.Close()
	t.Setenv("VAULT_TOKEN", "foo")
	v.Login()
	assert.Equal(t, "foo", v.client.Token())
}

func TestTokenLogin(t *testing.T) {
	server, v := MockServer(404, "Not Found")
	defer server.Close()
	t.Setenv("VAULT_TOKEN", "foo")

	token, err := v.TokenLogin()
	require.NoError(t, err)
	assert.Equal(t, "foo", token)
}
