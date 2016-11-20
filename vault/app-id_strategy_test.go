package vault

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAppIDAuthStrategy_NoEnvVars(t *testing.T) {
	auth, err := NewAppIDAuthStrategy(func(env string) string {
		return ""
	})
	assert.Error(t, err)
	assert.Nil(t, auth)
}

func TestNewAppIDAuthStrategy_NoVaultUserId(t *testing.T) {
	_, err := NewAppIDAuthStrategy(func(env string) string {
		if env == "VAULT_APP_ID" {
			return "foo"
		}
		return ""
	})
	assert.Error(t, err)
}

func TestNewAppIDAuthStrategy_NoVaultAppId(t *testing.T) {
	_, err := NewAppIDAuthStrategy(func(env string) string {
		if env == "VAULT_USER_ID" {
			return "bar"
		}
		return ""
	})
	assert.Error(t, err)
}

func TestNewAppIDAuthStrategy_Works(t *testing.T) {
	auth, err := NewAppIDAuthStrategy(func(env string) string {
		switch env {
		case "VAULT_APP_ID":
			return "foo"
		case "VAULT_USER_ID":
			return "bar"
		default:
			return ""
		}
	})
	assert.Nil(t, err)
	assert.Equal(t, "foo", auth.AppID)
	assert.Equal(t, "bar", auth.UserID)
}

func TestGetToken_AppIDErrorsGivenNetworkError(t *testing.T) {
	server, client := setupErrorHTTP()
	defer server.Close()

	vaultURL, _ := url.Parse("http://vault:8200")

	auth := &AppIDAuthStrategy{"foo", "bar", client}
	_, err := auth.GetToken(vaultURL)
	assert.Error(t, err)
}

func TestGetToken_AppIDErrorsGivenHTTPErrorStatus(t *testing.T) {
	server, client := setupHTTP(500, "application/json; charset=utf-8", `{}`)
	defer server.Close()

	vaultURL, _ := url.Parse("http://vault:8200")

	auth := &AppIDAuthStrategy{"foo", "bar", client}
	_, err := auth.GetToken(vaultURL)
	assert.Error(t, err)
}

func TestGetToken_AppIDErrorsGivenBadJSON(t *testing.T) {
	server, client := setupHTTP(200, "application/json; charset=utf-8", `{`)
	defer server.Close()

	vaultURL, _ := url.Parse("http://vault:8200")

	auth := &AppIDAuthStrategy{"foo", "bar", client}
	_, err := auth.GetToken(vaultURL)
	assert.Error(t, err)
}

func TestGetToken_AppID(t *testing.T) {
	server, client := setupHTTP(200, "application/json; charset=utf-8", `{"auth": {"client_token": "baz"}}`)
	defer server.Close()

	vaultURL, _ := url.Parse("http://vault:8200")

	auth := &AppIDAuthStrategy{"foo", "bar", client}
	token, err := auth.GetToken(vaultURL)
	assert.NoError(t, err)

	assert.Equal(t, "baz", token)
}

func setupHTTP(code int, mimetype string, body string) (*httptest.Server, *http.Client) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", mimetype)
		w.WriteHeader(code)
		fmt.Fprintln(w, body)
	}))

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				return url.Parse(server.URL)
			},
		},
	}

	return server, client
}

func setupErrorHTTP() (*httptest.Server, *http.Client) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boo")
	}))

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				return url.Parse(server.URL)
			},
		},
	}

	return server, client
}
