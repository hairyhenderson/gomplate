package vault

import (
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGitHubAuthStrategy(t *testing.T) {
	defer os.Unsetenv("VAULT_AUTH_GITHUB_TOKEN")
	defer os.Unsetenv("VAULT_AUTH_GITHUB_MOUNT")

	os.Unsetenv("VAULT_AUTH_GITHUB_TOKEN")
	assert.Nil(t, NewGitHubAuthStrategy())

	os.Setenv("VAULT_AUTH_GITHUB_TOKEN", "foo")
	auth := NewGitHubAuthStrategy()
	assert.Equal(t, "foo", auth.Token)
	assert.Equal(t, "github", auth.Mount)

	os.Setenv("VAULT_AUTH_GITHUB_MOUNT", "bar")
	auth = NewGitHubAuthStrategy()
	assert.Equal(t, "foo", auth.Token)
	assert.Equal(t, "bar", auth.Mount)
}

func TestGetToken_GitHubErrorsGivenNetworkError(t *testing.T) {
	server, client := setupErrorHTTP()
	defer server.Close()

	vaultURL, _ := url.Parse("http://vault:8200")

	auth := &GitHubAuthStrategy{"foo", "github", client}
	_, err := auth.GetToken(vaultURL)
	assert.Error(t, err)
}

func TestGetToken_GitHubErrorsGivenHTTPErrorStatus(t *testing.T) {
	server, client := setupHTTP(500, "application/json; charset=utf-8", `{}`)
	defer server.Close()

	vaultURL, _ := url.Parse("http://vault:8200")

	auth := &GitHubAuthStrategy{"foo", "github", client}
	_, err := auth.GetToken(vaultURL)
	assert.Error(t, err)
}

func TestGetToken_GitHubErrorsGivenBadJSON(t *testing.T) {
	server, client := setupHTTP(200, "application/json; charset=utf-8", `{`)
	defer server.Close()

	vaultURL, _ := url.Parse("http://vault:8200")

	auth := &GitHubAuthStrategy{"foo", "github", client}
	_, err := auth.GetToken(vaultURL)
	assert.Error(t, err)
}

func TestGetToken_GitHub(t *testing.T) {
	server, client := setupHTTP(200, "application/json; charset=utf-8", `{"auth": {"client_token": "baz"}}`)
	defer server.Close()

	vaultURL, _ := url.Parse("http://vault:8200")

	auth := &GitHubAuthStrategy{"foo", "github", client}
	token, err := auth.GetToken(vaultURL)
	assert.NoError(t, err)

	assert.Equal(t, "baz", token)
}
