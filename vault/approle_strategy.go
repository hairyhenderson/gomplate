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

// AppRoleAuthStrategy - an AuthStrategy that uses Vault's approle authentication backend.
type AppRoleAuthStrategy struct {
	RoleID   string `json:"role_id"`
	SecretID string `json:"secret_id"`
	Mount    string `json:"-"`
	hc       *http.Client
}

// NewAppRoleAuthStrategy - create an AuthStrategy that uses Vault's approle auth
// backend.
func NewAppRoleAuthStrategy() *AppRoleAuthStrategy {
	roleID := os.Getenv("VAULT_ROLE_ID")
	secretID := os.Getenv("VAULT_SECRET_ID")
	mount := os.Getenv("VAULT_AUTH_APPROLE_MOUNT")
	if mount == "" {
		mount = "approle"
	}
	if roleID != "" && secretID != "" {
		return &AppRoleAuthStrategy{roleID, secretID, mount, nil}
	}
	return nil
}

// GetHTTPClient configures the HTTP client with a timeout
func (a *AppRoleAuthStrategy) GetHTTPClient() *http.Client {
	if a.hc == nil {
		a.hc = &http.Client{Timeout: time.Second * 5}
	}
	return a.hc
}

// SetToken is a no-op for AppRoleAuthStrategy as a token hasn't been acquired yet
func (a *AppRoleAuthStrategy) SetToken(req *http.Request) {
	// no-op
}

// Do wraps http.Client.Do
func (a *AppRoleAuthStrategy) Do(req *http.Request) (*http.Response, error) {
	hc := a.GetHTTPClient()
	return hc.Do(req)
}

// GetToken - log in to the approle auth backend and return the client token
func (a *AppRoleAuthStrategy) GetToken(addr *url.URL) (string, error) {
	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(&a)

	u := &url.URL{}
	*u = *addr
	u.Path = "/v1/auth/" + a.Mount + "/login"
	res, err := requestAndFollow(a, "POST", u, buf.Bytes())
	if err != nil {
		return "", err
	}
	response := &AppRoleAuthResponse{}
	err = json.NewDecoder(res.Body).Decode(response)
	res.Body.Close()
	if err != nil {
		return "", err
	}
	if res.StatusCode != 200 {
		err := fmt.Errorf("Unexpected HTTP status %d on AppRole login to %s: %s", res.StatusCode, u, response)
		return "", err
	}
	return response.Auth.ClientToken, nil
}

// Revokable -
func (a *AppRoleAuthStrategy) Revokable() bool {
	return true
}

func (a *AppRoleAuthStrategy) String() string {
	return fmt.Sprintf("role_id: %s, secret_id: %s, mount: %s", a.RoleID, a.SecretID, a.Mount)
}

// AppRoleAuthResponse - the Auth response from /v1/auth/approle/login
type AppRoleAuthResponse struct {
	Auth struct {
		ClientToken   string   `json:"client_token"`
		LeaseDuration int64    `json:"lease_duration"`
		Metadata      struct{} `json:"metadata"`
		Policies      []string `json:"policies"`
		Renewable     bool     `json:"renewable"`
	} `json:"auth"`
}

func (a *AppRoleAuthResponse) String() string {
	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(&a)
	return string(buf.Bytes())
}
