package vault

import (
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUserPassAuthStrategy(t *testing.T) {
	os.Unsetenv("VAULT_AUTH_USERNAME")
	os.Unsetenv("VAULT_AUTH_PASSWORD")
	assert.Nil(t, NewUserPassAuthStrategy())

	os.Setenv("VAULT_AUTH_USERNAME", "foo")
	assert.Nil(t, NewUserPassAuthStrategy())

	os.Unsetenv("VAULT_AUTH_USERNAME")
	os.Setenv("VAULT_AUTH_PASSWORD", "bar")
	assert.Nil(t, NewUserPassAuthStrategy())

	os.Setenv("VAULT_AUTH_USERNAME", "foo")
	os.Setenv("VAULT_AUTH_PASSWORD", "bar")
	auth := NewUserPassAuthStrategy()
	assert.Equal(t, "foo", auth.Username)
	assert.Equal(t, "bar", auth.Password)
	assert.Equal(t, "userpass", auth.Mount)

	os.Setenv("VAULT_AUTH_USERNAME", "foo")
	os.Setenv("VAULT_AUTH_PASSWORD", "bar")
	os.Setenv("VAULT_AUTH_USERPASS_MOUNT", "baz")
	auth = NewUserPassAuthStrategy()
	assert.Equal(t, "foo", auth.Username)
	assert.Equal(t, "bar", auth.Password)
	assert.Equal(t, "baz", auth.Mount)
}

func TestGetToken_UserPassErrorsGivenNetworkError(t *testing.T) {
	server, client := setupErrorHTTP()
	defer server.Close()

	vaultURL, _ := url.Parse("http://vault:8200")

	auth := &UserPassAuthStrategy{"foo", "bar", "userpass", client}
	_, err := auth.GetToken(vaultURL)
	assert.Error(t, err)
}

func TestGetToken_UserPassErrorsGivenHTTPErrorStatus(t *testing.T) {
	server, client := setupHTTP(500, "application/json; charset=utf-8", `{}`)
	defer server.Close()

	vaultURL, _ := url.Parse("http://vault:8200")

	auth := &UserPassAuthStrategy{"foo", "bar", "userpass", client}
	_, err := auth.GetToken(vaultURL)
	assert.Error(t, err)
}

func TestGetToken_UserPassErrorsGivenBadJSON(t *testing.T) {
	server, client := setupHTTP(200, "application/json; charset=utf-8", `{`)
	defer server.Close()

	vaultURL, _ := url.Parse("http://vault:8200")

	auth := &UserPassAuthStrategy{"foo", "bar", "userpass", client}
	_, err := auth.GetToken(vaultURL)
	assert.Error(t, err)
}

func TestGetToken_UserPass(t *testing.T) {
	server, client := setupHTTP(200, "application/json; charset=utf-8", `{"auth": {"client_token": "baz"}}`)
	defer server.Close()

	vaultURL, _ := url.Parse("http://vault:8200")

	auth := &UserPassAuthStrategy{"foo", "bar", "userpass", client}
	token, err := auth.GetToken(vaultURL)
	assert.NoError(t, err)

	assert.Equal(t, "baz", token)
}
