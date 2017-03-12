package vault

import (
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAppRoleAuthStrategy(t *testing.T) {
	defer os.Unsetenv("VAULT_ROLE_ID")
	defer os.Unsetenv("VAULT_SECRET_ID")
	defer os.Unsetenv("VAULT_AUTH_APPROLE_MOUNT")

	os.Unsetenv("VAULT_ROLE_ID")
	os.Unsetenv("VAULT_SECRET_ID")
	assert.Nil(t, NewAppRoleAuthStrategy())

	os.Setenv("VAULT_ROLE_ID", "foo")
	assert.Nil(t, NewAppRoleAuthStrategy())

	os.Unsetenv("VAULT_ROLE_ID")
	os.Setenv("VAULT_SECRET_ID", "bar")
	assert.Nil(t, NewAppRoleAuthStrategy())

	os.Setenv("VAULT_ROLE_ID", "foo")
	os.Setenv("VAULT_SECRET_ID", "bar")
	auth := NewAppRoleAuthStrategy()
	assert.Equal(t, "foo", auth.RoleID)
	assert.Equal(t, "bar", auth.SecretID)
	assert.Equal(t, "approle", auth.Mount)

	os.Setenv("VAULT_ROLE_ID", "baz")
	os.Setenv("VAULT_SECRET_ID", "qux")
	os.Setenv("VAULT_AUTH_APPROLE_MOUNT", "quux")
	auth = NewAppRoleAuthStrategy()
	assert.Equal(t, "baz", auth.RoleID)
	assert.Equal(t, "qux", auth.SecretID)
	assert.Equal(t, "quux", auth.Mount)
}

func TestGetToken_AppRoleErrorsGivenNetworkError(t *testing.T) {
	server, client := setupErrorHTTP()
	defer server.Close()

	vaultURL, _ := url.Parse("http://vault:8200")

	auth := &AppRoleAuthStrategy{"foo", "bar", "approle", client}
	_, err := auth.GetToken(vaultURL)
	assert.Error(t, err)
}

func TestGetToken_AppRoleErrorsGivenHTTPErrorStatus(t *testing.T) {
	server, client := setupHTTP(500, "application/json; charset=utf-8", `{}`)
	defer server.Close()

	vaultURL, _ := url.Parse("http://vault:8200")

	auth := &AppRoleAuthStrategy{"foo", "bar", "approle", client}
	_, err := auth.GetToken(vaultURL)
	assert.Error(t, err)
}

func TestGetToken_AppRoleErrorsGivenBadJSON(t *testing.T) {
	server, client := setupHTTP(200, "application/json; charset=utf-8", `{`)
	defer server.Close()

	vaultURL, _ := url.Parse("http://vault:8200")

	auth := &AppRoleAuthStrategy{"foo", "bar", "approle", client}
	_, err := auth.GetToken(vaultURL)
	assert.Error(t, err)
}

func TestGetToken_AppRole(t *testing.T) {
	server, client := setupHTTP(200, "application/json; charset=utf-8", `{"auth": {"client_token": "baz"}}`)
	defer server.Close()

	vaultURL, _ := url.Parse("http://vault:8200")

	auth := &AppRoleAuthStrategy{"foo", "bar", "approle", client}
	token, err := auth.GetToken(vaultURL)
	assert.NoError(t, err)

	assert.Equal(t, "baz", token)
}
