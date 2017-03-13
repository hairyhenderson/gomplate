package vault

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"
)

// AuthStrategy -
type AuthStrategy interface {
	fmt.Stringer
	GetToken(addr *url.URL) (string, error)
	RevokeToken(addr *url.URL) error
}

func getAuthStrategy() AuthStrategy {
	if auth := NewAppRoleAuthStrategy(); auth != nil {
		return auth
	}
	if auth := NewAppIDAuthStrategy(); auth != nil {
		return auth
	}
	if auth := NewGitHubAuthStrategy(); auth != nil {
		return auth
	}
	if auth := NewUserPassAuthStrategy(); auth != nil {
		return auth
	}
	if auth := NewTokenAuthStrategy(); auth != nil {
		return auth
	}
	logFatal("No vault auth strategy configured")
	return nil
}

// Strategy -
type Strategy struct {
	Mount string `json:"-"`
	hc    *http.Client
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

// GetPath -
func (a *Strategy) GetPath() string {
	return "/v1/auth/" + a.Mount + "/login"
}

// GetToken - log in to the app-id auth backend and return the client token
func (a *Strategy) GetToken(addr *url.URL) (string, error) {
	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(&a)

	u := &url.URL{}
	*u = *addr
	u.Path = a.GetPath()
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

// RevokeToken - revoke the current auth token - effectively logging out
func (a *Strategy) RevokeToken(addr *url.URL) error {
	u := &url.URL{}
	*u = *addr
	u.Path = "/v1/auth/token/revoke-self"

	res, err := requestAndFollow(a, "POST", u, nil)

	if err != nil {
		return err
	}

	if res.StatusCode != 204 {
		b, _ := ioutil.ReadAll(res.Body)
		return errors.Errorf("Unexpected HTTP status %d on RevokeToken from %s (body was: %s)", res.StatusCode, u, string(b))
	}
	return nil
}

// AuthResponse - the Auth response from /v1/auth/app-id/login
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
