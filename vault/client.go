package vault

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"
)

// logFatal is defined so log.Fatal calls can be overridden for testing
var logFatal = log.Fatal

// Client -
type Client struct {
	Addr *url.URL
	Auth AuthStrategy
	// The cached auth token
	token string
	hc    *http.Client
}

// AuthStrategy -
type AuthStrategy interface {
	fmt.Stringer
	GetToken(addr *url.URL) (string, error)
	Revokable() bool
}

// NewClient - instantiate a new
func NewClient() *Client {
	u := getVaultAddr()
	auth := getAuthStrategy()
	return &Client{u, auth, "", nil}
}

func getVaultAddr() *url.URL {
	vu := os.Getenv("VAULT_ADDR")
	if vu == "" {
		logFatal("VAULT_ADDR is an unparseable URL!")
		return nil
	}
	u, err := url.Parse(vu)
	if err != nil {
		logFatal("VAULT_ADDR is an unparseable URL!", err)
		return nil
	}
	return u
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

// GetHTTPClient returns a client configured w/X-Vault-Token header
func (c *Client) GetHTTPClient() *http.Client {
	if c.hc == nil {
		c.hc = &http.Client{
			Timeout: time.Second * 5,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				c.SetToken(req)
				return nil
			},
		}
	}
	return c.hc
}

// SetToken adds an X-Vault-Token header to the request
func (c *Client) SetToken(req *http.Request) {
	req.Header.Set("X-Vault-Token", c.token)
}

// Do wraps http.Client.Do
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	hc := c.GetHTTPClient()
	return hc.Do(req)
}

// Login - log in to Vault with the discovered auth backend and save the token
func (c *Client) Login() error {
	token, err := c.Auth.GetToken(c.Addr)
	if err != nil {
		logFatal(err)
		return err
	}
	c.token = token
	return nil
}

// RevokeToken - revoke the current auth token - effectively logging out
func (c *Client) RevokeToken() {
	// only do it if the auth strategy supports it!
	if !c.Auth.Revokable() {
		return
	}

	u := &url.URL{}
	*u = *c.Addr
	u.Path = "/v1/auth/token/revoke-self"

	res, err := requestAndFollow(c, "POST", u, nil)

	if err != nil {
		log.Println("Error while revoking Vault Token", err)
	}

	if res.StatusCode != 204 {
		log.Printf("Unexpected HTTP status %d on RevokeToken from %s (token was %s)", res.StatusCode, u, c.token)
	}
}

func (c *Client) Read(path string) ([]byte, error) {
	path = normalizeURLPath(path)

	u := &url.URL{}
	*u = *c.Addr
	u.Path = "/v1" + path

	res, err := requestAndFollow(c, "GET", u, nil)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		err = fmt.Errorf("Unexpected HTTP status %d on Read from %s: %s", res.StatusCode, path, body)
		return nil, err
	}

	response := make(map[string]interface{})
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Println("argh - couldn't decode the response", err)
		return nil, err
	}

	if _, ok := response["data"]; !ok {
		return nil, fmt.Errorf("Unexpected HTTP body on Read for %s: %s", path, body)
	}

	return json.Marshal(response["data"])
}

var rxDupSlashes = regexp.MustCompile(`/{2,}`)

func normalizeURLPath(path string) string {
	if len(path) > 0 {
		path = rxDupSlashes.ReplaceAllString(path, "/")
	}
	return path
}

// ReadResponse -
type ReadResponse struct {
	Data struct {
		Value string `json:"value"`
	} `json:"data"`
}
