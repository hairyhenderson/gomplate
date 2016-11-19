package vault

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testClient struct{}

func (tc *testClient) GetHTTPClient() *http.Client {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqStr := fmt.Sprintf("%s %s", r.Method, r.URL)
		switch reqStr {
		case "POST http://vaultA:8500/v1/foo":
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Location", "http://vaultB:8500/v1/foo")
			w.WriteHeader(http.StatusTemporaryRedirect)
			fmt.Fprintln(w, "{}")
		case "POST http://vaultB:8500/v1/foo":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "{}")
		default:
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "{ 'message': 'Unexpected request: %s'}", reqStr)
		}
	}))
	return &http.Client{
		Transport: &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				return url.Parse(server.URL)
			},
		},
	}
}

func (tc *testClient) SetToken(req *http.Request) {
	req.Header.Set("X-Vault-Token", "dead-beef-cafe-babe")
}

func (tc *testClient) Do(req *http.Request) (*http.Response, error) {
	hc := tc.GetHTTPClient()
	return hc.Do(req)
}

func TestRequestAndFollow_GetWithRedirect(t *testing.T) {
	tc := &testClient{}
	u, _ := url.Parse("http://vaultA:8500/v1/foo")

	res, err := requestAndFollow(tc, "POST", u, nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

}

func TestRequestAndFollow_GetNoRedirect(t *testing.T) {
	tc := &testClient{}
	u, _ := url.Parse("http://vaultB:8500/v1/foo")

	res, err := requestAndFollow(tc, "POST", u, nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}
