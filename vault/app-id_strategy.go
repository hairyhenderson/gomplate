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
	hc     *http.Client
}

// NewAppIDAuthStrategy - create an AuthStrategy that uses Vault's app-id auth
// backend.
func NewAppIDAuthStrategy() *AppIDAuthStrategy {
	appID := os.Getenv("VAULT_APP_ID")
	userID := os.Getenv("VAULT_USER_ID")
	if appID != "" && userID != "" {
		return &AppIDAuthStrategy{appID, userID, nil}
	}
	return nil
}

// GetToken - log in to the app-id auth backend and return the client token
func (a *AppIDAuthStrategy) GetToken(addr *url.URL) (string, error) {
	if a.hc == nil {
		a.hc = &http.Client{Timeout: time.Second * 5}
	}
	client := a.hc

	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(&a)

	u := &url.URL{}
	*u = *addr
	u.Path = "/v1/auth/app-id/login"
	res, err := client.Post(u.String(), "application/json; charset=utf-8", buf)
	if err != nil {
		return "", err
	}
	response := &AuthResponse{}
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
	return fmt.Sprintf("app-id: %s, user-id: %s", a.AppID, a.UserID)
}

// AuthResponse - the Auth response from /v1/auth/app-id/login
type AuthResponse struct {
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

func (a *AuthResponse) String() string {
	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(&a)
	return string(buf.Bytes())
}
