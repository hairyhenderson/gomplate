package vault

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogin(t *testing.T) {
	server, v := MockServer(404, "Not Found")
	defer server.Close()
	os.Setenv("VAULT_TOKEN", "foo")
	defer os.Unsetenv("VAULT_TOKEN")
	v.Login()
	assert.Equal(t, "foo", v.client.Token())
}

func TestTokenLogin(t *testing.T) {
	server, v := MockServer(404, "Not Found")
	defer server.Close()
	os.Setenv("VAULT_TOKEN", "foo")
	defer os.Unsetenv("VAULT_TOKEN")

	token, err := v.TokenLogin()
	assert.NoError(t, err)
	assert.Equal(t, "foo", token)
}
