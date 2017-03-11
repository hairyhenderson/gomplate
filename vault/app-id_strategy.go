package vault

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
)

// AppIDAuthStrategy - an AuthStrategy that uses Vault's app-id authentication backend.
type AppIDAuthStrategy struct {
	AppID  string `json:"app_id"`
	UserID string `json:"user_id"`
	Mount  string `json:"-"`
	hc     *http.Client
}

// NewAppIDAuthStrategy - create an AuthStrategy that uses Vault's app-id auth
// backend.
func NewAppIDAuthStrategy() *AppIDAuthStrategy {
	appID := os.Getenv("VAULT_APP_ID")
	userID := os.Getenv("VAULT_USER_ID")
	mount := os.Getenv("VAULT_AUTH_APP_ID_MOUNT")
	if mount == "" {
		mount = "app-id"
	}
	if appID != "" && userID != "" {
		return &AppIDAuthStrategy{appID, userID, mount, nil}
	}
	return nil
}

// GetHTTPClient configures the HTTP client with a timeout
func (a *AppIDAuthStrategy) GetHTTPClient() *http.Client {
	if a.hc == nil {
		a.hc = &http.Client{Timeout: time.Second * 5}
	}
	return a.hc
}

// SetToken is a no-op for AppIDAuthStrategy as a token hasn't been acquired yet
func (a *AppIDAuthStrategy) SetToken(req *http.Request) {
	// no-op
}

// Do wraps http.Client.Do
func (a *AppIDAuthStrategy) Do(req *http.Request) (*http.Response, error) {
	hc := a.GetHTTPClient()
	return hc.Do(req)
}

// GetToken - log in to the app-id auth backend and return the client token
func (a *AppIDAuthStrategy) GetToken(addr *url.URL) (string, error) {
	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(&a)

	u := &url.URL{}
	*u = *addr
	u.Path = "/v1/auth/" + a.Mount + "/login"
	res, err := requestAndFollow(a, "POST", u, buf.Bytes())
	if err != nil {
		return "", err
	}
	response := &AppIDAuthResponse{}
	err = json.NewDecoder(res.Body).Decode(response)
	res.Body.Close()
	if err != nil {
		return "", err
	}
	if res.StatusCode != 200 {
		err := fmt.Errorf("Unexpected HTTP status %d on AppId login to %s: %s", res.StatusCode, u, response)
		return "", err
	}
	return response.Auth.ClientToken, nil
}

// Revokable -
func (a *AppIDAuthStrategy) Revokable() bool {
	return true
}

func (a *AppIDAuthStrategy) String() string {
	return fmt.Sprintf("app-id: %s, user-id: %s, mount: %s", a.AppID, a.UserID, a.Mount)
}

// AppIDAuthResponse - the Auth response from /v1/auth/app-id/login
type AppIDAuthResponse struct {
	Auth struct {
		ClientToken   string `json:"client_token"`
		LeaseDuration int64  `json:"lease_duration"`
		Metadata      struct {
			AppID  string `json:"app-id"`
			UserID string `json:"user-id"`
		} `json:"metadata"`
		Policies  []string `json:"policies"`
		Renewable bool     `json:"renewable"`
	} `json:"auth"`
}

func (a *AppIDAuthResponse) String() string {
	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(&a)
	return string(buf.Bytes())
}
