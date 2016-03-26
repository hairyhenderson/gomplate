package aws

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
)

// MockServer -
func MockServer(code int, body string) (*httptest.Server, *Ec2Meta) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
		fmt.Fprintln(w, body)
	}))

	tr := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}
	httpClient := &http.Client{Transport: tr}

	client := &Ec2Meta{server.URL + "/", httpClient}
	return server, client
}
