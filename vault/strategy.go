package vault

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/blang/vfs"
)

// AuthStrategy -
type AuthStrategy interface {
	GetToken(addr *url.URL) (string, error)
	Revokable() bool
}

func getAuthStrategy(fsOverrides ...vfs.Filesystem) AuthStrategy {
	var fs vfs.Filesystem
	if len(fsOverrides) == 0 {
		fs = vfs.OS()
	} else {
		fs = fsOverrides[0]
	}

	if auth := AppRoleStrategy(fs); auth != nil {
		return auth
	}
	if auth := AppIDStrategy(fs); auth != nil {
		return auth
	}
	if auth := GitHubStrategy(fs); auth != nil {
		return auth
	}
	if auth := UserPassStrategy(fs); auth != nil {
		return auth
	}
	if auth := NewTokenStrategy(fs); auth != nil {
		return auth
	}
	logFatal("No vault auth strategy configured")
	return nil
}

// Strategy - an auth strategy configured through the Vault API
type Strategy struct {
	mount string
	hc    *http.Client
	path  string
	body  map[string]string
}

// GetHTTPClient configures the HTTP client with a timeout
func (a *Strategy) GetHTTPClient() *http.Client {
	if a.hc == nil {
		a.hc = &http.Client{Timeout: time.Second * 5}
	}
	return a.hc
}

// SetToken is a no-op here as a token hasn't been acquired yet
func (a *Strategy) SetToken(req *http.Request) {
	// no-op
}

// Do wraps http.Client.Do
func (a *Strategy) Do(req *http.Request) (*http.Response, error) {
	hc := a.GetHTTPClient()
	return hc.Do(req)
}

// GetToken - log in to the app-id auth backend and return the client token
func (a *Strategy) GetToken(addr *url.URL) (string, error) {
	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(&a.body)

	u := &url.URL{}
	*u = *addr
	if a.path == "" {
		a.path = "/v1/auth/" + a.mount + "/login"
	}
	u.Path = a.path
	res, err := requestAndFollow(a, "POST", u, buf.Bytes())
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
func (a *Strategy) Revokable() bool {
	return true
}

// AuthResponse - the Auth response
type AuthResponse struct {
	Auth struct {
		ClientToken   string   `json:"client_token"`
		LeaseDuration int64    `json:"lease_duration"`
		Metadata      struct{} `json:"metadata"`
		Policies      []string `json:"policies"`
		Renewable     bool     `json:"renewable"`
	} `json:"auth"`
}

func (a *AuthResponse) String() string {
	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(&a)
	return string(buf.Bytes())
}
