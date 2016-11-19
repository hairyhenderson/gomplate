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

// GitHubAuthStrategy - an AuthStrategy that uses Vault's app-id authentication backend.
type GitHubAuthStrategy struct {
	Token string `json:"token"`
	Mount string `json:"-"`
	hc    *http.Client
}

// NewGitHubAuthStrategy - create an AuthStrategy that uses Vault's app-id auth
// backend.
func NewGitHubAuthStrategy() *GitHubAuthStrategy {
	mount := os.Getenv("VAULT_AUTH_GITHUB_MOUNT")
	if mount == "" {
		mount = "github"
	}
	token := os.Getenv("VAULT_AUTH_GITHUB_TOKEN")
	if token != "" {
		return &GitHubAuthStrategy{token, mount, nil}
	}
	return nil
}

// GetHTTPClient configures the HTTP client with a timeout
func (a *GitHubAuthStrategy) GetHTTPClient() *http.Client {
	if a.hc == nil {
		a.hc = &http.Client{Timeout: time.Second * 5}
	}
	return a.hc
}

// SetToken is a no-op for GitHubAuthStrategy as a token hasn't been acquired yet
func (a *GitHubAuthStrategy) SetToken(req *http.Request) {
	// no-op
}

// Do wraps http.Client.Do
func (a *GitHubAuthStrategy) Do(req *http.Request) (*http.Response, error) {
	hc := a.GetHTTPClient()
	return hc.Do(req)
}

// GetToken - log in to the auth backend and return the client token
func (a *GitHubAuthStrategy) GetToken(addr *url.URL) (string, error) {
	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(&a)

	u := &url.URL{}
	*u = *addr
	u.Path = "/v1/auth/" + a.Mount + "/login"
	res, err := requestAndFollow(a, "POST", u, buf.Bytes())
	if err != nil {
		return "", err
	}
	response := &GitHubAuthResponse{}
	err = json.NewDecoder(res.Body).Decode(response)
	res.Body.Close()
	if err != nil {
		return "", err
	}
	if res.StatusCode != 200 {
		err := fmt.Errorf("Unexpected HTTP status %d on GitHub login to %s: %s", res.StatusCode, u, response)
		return "", err
	}
	return response.Auth.ClientToken, nil
}

// Revokable -
func (a *GitHubAuthStrategy) Revokable() bool {
	return true
}

func (a *GitHubAuthStrategy) String() string {
	return fmt.Sprintf("token: %s, mount: %s", a.Token, a.Mount)
}

// GitHubAuthResponse - the Auth response from /v1/auth/app-id/login
type GitHubAuthResponse struct {
	Auth struct {
		ClientToken   string `json:"client_token"`
		LeaseDuration int64  `json:"lease_duration"`
		Metadata      struct {
			Username string `json:"username"`
			Org      string `json:"org"`
		} `json:"metadata"`
		Policies  []string `json:"policies"`
		Renewable bool     `json:"renewable"`
	} `json:"auth"`
}

func (a *GitHubAuthResponse) String() string {
	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(&a)
	return string(buf.Bytes())
}
