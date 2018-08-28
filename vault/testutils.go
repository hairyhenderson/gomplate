package vault

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/hashicorp/vault/api"
)

// MockServer -
func MockServer(code int, body string) (*httptest.Server, *Vault) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
		// nolint: errcheck
		fmt.Fprintln(w, body)
	}))

	tr := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}
	httpClient := &http.Client{Transport: tr}
	config := &api.Config{
		Address:    server.URL,
		HttpClient: httpClient,
	}
	// nolint: gosec
	c, _ := api.NewClient(config)
	return server, &Vault{c}
}
