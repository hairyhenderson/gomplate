package vault

import (
	"log"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var spyLogFatalMsg string

func restoreLogFatal() {
	logFatal = log.Fatal
}

func mockLogFatal(args ...interface{}) {
	spyLogFatalMsg = (args[0]).(string)
}

func setupMockLogFatal() {
	logFatal = mockLogFatal
	spyLogFatalMsg = ""
}
func TestNewClient_NoVaultAddr(t *testing.T) {
	os.Unsetenv("VAULT_ADDR")
	defer restoreLogFatal()
	setupMockLogFatal()
	c := NewClient()
	assert.Nil(t, c.Addr)
	assert.Equal(t, "VAULT_ADDR is an unparseable URL!", spyLogFatalMsg)
}

func TestLogin_NoAuthStrategy(t *testing.T) {
	os.Setenv("VAULT_ADDR", "https://localhost:8500")
	os.Unsetenv("VAULT_APP_ID")
	os.Unsetenv("VAULT_USER_ID")
	os.Setenv("HOME", "/tmp")
	defer restoreLogFatal()
	setupMockLogFatal()
	_ = NewClient()
	assert.Equal(t, "No vault auth strategy configured", spyLogFatalMsg)
}

func TestLogin_SavesToken(t *testing.T) {
	auth := &TokenAuthStrategy{"foo"}
	client := &Client{
		Auth: auth,
	}
	err := client.Login()
	assert.NoError(t, err)
	assert.Equal(t, "foo", client.token)
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
