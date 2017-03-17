package vault

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetHTTPClient(t *testing.T) {
	strat := &Strategy{}
	assert.IsType(t, &http.Client{}, strat.GetHTTPClient())

	expected := &http.Client{Timeout: time.Second * 1}
	strat = &Strategy{mount: "", hc: expected}
	assert.Equal(t, expected, strat.GetHTTPClient())
}

func TestGetToken_ErrorsGivenNetworkError(t *testing.T) {
	server, client := setupErrorHTTP()
	defer server.Close()

	vaultURL, _ := url.Parse("http://vault:8200")

	auth := &Strategy{hc: client}
	_, err := auth.GetToken(vaultURL)
	assert.Error(t, err)
}

func TestGetToken_ErrorsGivenHTTPErrorStatus(t *testing.T) {
	server, client := setupHTTP(500, "application/json; charset=utf-8", `{}`)
	defer server.Close()

	vaultURL, _ := url.Parse("http://vault:8200")

	auth := &Strategy{hc: client}
	_, err := auth.GetToken(vaultURL)
	assert.Error(t, err)
}

func TestGetToken_ErrorsGivenBadJSON(t *testing.T) {
	server, client := setupHTTP(200, "application/json; charset=utf-8", `{`)
	defer server.Close()

	vaultURL, _ := url.Parse("http://vault:8200")

	auth := &Strategy{hc: client}
	_, err := auth.GetToken(vaultURL)
	assert.Error(t, err)
}

func TestGetToken(t *testing.T) {
	server, client := setupHTTP(200, "application/json; charset=utf-8", `{"auth": {"client_token": "baz"}}`)
	defer server.Close()

	vaultURL, _ := url.Parse("http://vault:8200")

	auth := &Strategy{hc: client}
	token, err := auth.GetToken(vaultURL)
	assert.NoError(t, err)

	assert.Equal(t, "baz", token)
}
