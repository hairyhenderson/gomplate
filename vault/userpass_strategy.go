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

// UserPassAuthStrategy - an AuthStrategy that uses Vault's userpass authentication backend.
type UserPassAuthStrategy struct {
	Username string `json:"-"`
	Password string `json:"password"`
	Mount    string `json:"-"`
	hc       *http.Client
}

// NewUserPassAuthStrategy - create an AuthStrategy that uses Vault's userpass auth
// backend.
func NewUserPassAuthStrategy() *UserPassAuthStrategy {
	username := os.Getenv("VAULT_AUTH_USERNAME")
	password := os.Getenv("VAULT_AUTH_PASSWORD")
	mount := os.Getenv("VAULT_AUTH_USERPASS_MOUNT")
	if mount == "" {
		mount = "userpass"
	}
	if username != "" && password != "" {
		return &UserPassAuthStrategy{username, password, mount, nil}
	}
	return nil
}

// GetHTTPClient configures the HTTP client with a timeout
func (a *UserPassAuthStrategy) GetHTTPClient() *http.Client {
	if a.hc == nil {
		a.hc = &http.Client{Timeout: time.Second * 5}
	}
	return a.hc
}

// SetToken is a no-op for UserPassAuthStrategy as a token hasn't been acquired yet
func (a *UserPassAuthStrategy) SetToken(req *http.Request) {
	// no-op
}

// Do wraps http.Client.Do
func (a *UserPassAuthStrategy) Do(req *http.Request) (*http.Response, error) {
	hc := a.GetHTTPClient()
	return hc.Do(req)
}

// GetToken - log in to the app-id auth backend and return the client token
func (a *UserPassAuthStrategy) GetToken(addr *url.URL) (string, error) {
	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(&a)

	u := &url.URL{}
	*u = *addr
	u.Path = "/v1/auth/" + a.Mount + "/login/" + a.Username
	res, err := requestAndFollow(a, "POST", u, buf.Bytes())
	if err != nil {
		return "", err
	}
	response := &UserPassAuthResponse{}
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
func (a *UserPassAuthStrategy) Revokable() bool {
	return true
}

func (a *UserPassAuthStrategy) String() string {
	return fmt.Sprintf("username: %s, password: %s, mount: %s", a.Username, a.Password, a.Mount)
}

// UserPassAuthResponse - the Auth response from /v1/auth/username/login
type UserPassAuthResponse struct {
	Auth struct {
		ClientToken   string `json:"client_token"`
		LeaseDuration int64  `json:"lease_duration"`
		Metadata      struct {
			Username string `json:"username"`
		} `json:"metadata"`
		Policies  []string `json:"policies"`
		Renewable bool     `json:"renewable"`
	} `json:"auth"`
}

func (a *UserPassAuthResponse) String() string {
	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(&a)
	return string(buf.Bytes())
}
