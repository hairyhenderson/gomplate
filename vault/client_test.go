package vault

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogin_SavesToken(t *testing.T) {
	auth := &TokenAuthStrategy{"foo"}
	client := &Client{
		Auth: auth,
	}
	err := client.Login()
	assert.NoError(t, err)
	assert.Equal(t, "foo", client.token)
}

func TestNewClient_FailsWithNoEnvVars(t *testing.T) {
	_, err := NewClientUseGetenv(func(e string) string {
		return ""
	})
	assert.Error(t, err)
}

func TestRead_ErrorsGivenNetworkError(t *testing.T) {
	server, hc := setupErrorHTTP()
	defer server.Close()

	auth := &TokenAuthStrategy{"foo"}
	vaultURL, _ := url.Parse("http://vault:8200")
	client := &Client{
		Addr:  vaultURL,
		Auth:  auth,
		token: "foo",
		hc:    hc,
	}
	_, err := client.Read("secret/bar")
	assert.Error(t, err)
}

func TestRead_ErrorsGivenNonOKStatus(t *testing.T) {
	server, hc := setupHTTP(404, "application/json; charset=utf-8", `{}`)
	defer server.Close()

	auth := &TokenAuthStrategy{"foo"}
	vaultURL, _ := url.Parse("http://vault:8200")
	client := &Client{
		Addr:  vaultURL,
		Auth:  auth,
		token: "foo",
		hc:    hc,
	}
	_, err := client.Read("secret/bar")
	assert.Error(t, err)
}

func TestRead_ErrorsGivenBadJSON(t *testing.T) {
	server, hc := setupHTTP(200, "application/json; charset=utf-8", `{`)
	defer server.Close()

	auth := &TokenAuthStrategy{"foo"}
	vaultURL, _ := url.Parse("http://vault:8200")
	client := &Client{
		Addr:  vaultURL,
		Auth:  auth,
		token: "foo",
		hc:    hc,
	}
	_, err := client.Read("secret/bar")
	assert.Error(t, err)
}

func TestRead_ErrorsGivenWrongJSON(t *testing.T) {
	server, hc := setupHTTP(200, "application/json; charset=utf-8", `{}`)
	defer server.Close()

	auth := &TokenAuthStrategy{"foo"}
	vaultURL, _ := url.Parse("http://vault:8200")
	client := &Client{
		Addr:  vaultURL,
		Auth:  auth,
		token: "foo",
		hc:    hc,
	}
	_, err := client.Read("secret/bar")
	assert.Error(t, err)
}

func TestRead_ReturnsDataProp(t *testing.T) {
	server, hc := setupHTTP(200, "application/json; charset=utf-8", `{"data": {"value": "hi"}}`)
	defer server.Close()

	auth := &TokenAuthStrategy{"foo"}
	vaultURL, _ := url.Parse("http://vault:8200")
	client := &Client{
		Addr:  vaultURL,
		Auth:  auth,
		token: "foo",
		hc:    hc,
	}
	value, err := client.Read("secret/bar")
	assert.NoError(t, err)
	assert.Equal(t, []byte(`{"value":"hi"}`), value)
}

type fakeAuth struct {
	revokable bool
	token     string
}

func (a *fakeAuth) String() string {
	return a.token
}

func (a *fakeAuth) GetToken(addr *url.URL) (string, error) {
	return a.token, nil
}

func (a *fakeAuth) Revokable() bool {
	return a.revokable
}

func TestRevokeToken_NoopGivenNonRevokableAuth(t *testing.T) {
	auth := &fakeAuth{false, "foo"}
	client := &Client{
		Auth: auth,
	}
	client.Login()
	client.RevokeToken()
	assert.Equal(t, "foo", client.token)
}

func TestRevokeToken(t *testing.T) {
	server, hc := setupHTTP(204, "application/json; charset=utf-8", ``)
	defer server.Close()

	auth := &fakeAuth{true, "foo"}
	vaultURL, _ := url.Parse("http://vault:8200")
	client := &Client{
		Addr:  vaultURL,
		Auth:  auth,
		token: "foo",
		hc:    hc,
	}
	client.RevokeToken()
}
