package vault

import (
	"bytes"
	"net/http"
	"net/url"
)

// httpClient
type httpClient interface {
	GetHTTPClient() *http.Client
	SetToken(req *http.Request)
	Do(req *http.Request) (*http.Response, error)
}

func requestAndFollow(hc httpClient, method string, u *url.URL, body []byte) (*http.Response, error) {
	var res *http.Response
	var err error
	for attempts := 0; attempts < 2; attempts++ {
		reader := bytes.NewReader(body)
		req, err := http.NewRequest(method, u.String(), reader)

		if err != nil {
			return nil, err
		}
		hc.SetToken(req)
		if method == "POST" {
			req.Header.Set("Content-Type", "application/json; charset=utf-8")
		}

		res, err = hc.Do(req)
		if err != nil {
			return nil, err
		}
		if res.StatusCode == http.StatusTemporaryRedirect {
			res.Body.Close()
			location, errLocation := res.Location()
			if errLocation != nil {
				return nil, errLocation
			}
			u.Host = location.Host
		} else {
			break
		}
	}
	return res, err
}
