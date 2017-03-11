package vault

import (
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAppIDAuthStrategy(t *testing.T) {
	os.Unsetenv("VAULT_APP_ID")
	os.Unsetenv("VAULT_USER_ID")
	assert.Nil(t, NewAppIDAuthStrategy())

	os.Setenv("VAULT_APP_ID", "foo")
	assert.Nil(t, NewAppIDAuthStrategy())

	os.Unsetenv("VAULT_APP_ID")
	os.Setenv("VAULT_USER_ID", "bar")
	assert.Nil(t, NewAppIDAuthStrategy())

	os.Setenv("VAULT_APP_ID", "foo")
	os.Setenv("VAULT_USER_ID", "bar")
	auth := NewAppIDAuthStrategy()
	assert.Equal(t, "foo", auth.AppID)
	assert.Equal(t, "bar", auth.UserID)
}

func TestGetToken_AppIDErrorsGivenNetworkError(t *testing.T) {
	server, client := setupErrorHTTP()
	defer server.Close()

	vaultURL, _ := url.Parse("http://vault:8200")

	auth := &AppIDAuthStrategy{"foo", "bar", "app-id", client}
	_, err := auth.GetToken(vaultURL)
	assert.Error(t, err)
}

func TestGetToken_AppIDErrorsGivenHTTPErrorStatus(t *testing.T) {
	server, client := setupHTTP(500, "application/json; charset=utf-8", `{}`)
	defer server.Close()

	vaultURL, _ := url.Parse("http://vault:8200")

	auth := &AppIDAuthStrategy{"foo", "bar", "app-id", client}
	_, err := auth.GetToken(vaultURL)
	assert.Error(t, err)
}

func TestGetToken_AppIDErrorsGivenBadJSON(t *testing.T) {
	server, client := setupHTTP(200, "application/json; charset=utf-8", `{`)
	defer server.Close()

	vaultURL, _ := url.Parse("http://vault:8200")

	auth := &AppIDAuthStrategy{"foo", "bar", "app-id", client}
	_, err := auth.GetToken(vaultURL)
	assert.Error(t, err)
}

func TestGetToken_AppID(t *testing.T) {
	server, client := setupHTTP(200, "application/json; charset=utf-8", `{"auth": {"client_token": "baz"}}`)
	defer server.Close()

	vaultURL, _ := url.Parse("http://vault:8200")

	auth := &AppIDAuthStrategy{"foo", "bar", "app-id", client}
	token, err := auth.GetToken(vaultURL)
	assert.NoError(t, err)

	assert.Equal(t, "baz", token)
}
