package datasources

import (
	"context"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"time"

	"github.com/hairyhenderson/gomplate/v3/internal/config"
)

type httpRequester struct {
	hc *http.Client
}

func (r *httpRequester) Request(ctx context.Context, u *url.URL, header http.Header) (*Response, error) {
	if r.hc == nil {
		r.hc = config.HTTPClientFromContext(ctx)
		if r.hc == http.DefaultClient {
			r.hc = &http.Client{Timeout: time.Second * 5}
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header = header

	res, err := r.hc.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		defer res.Body.Close()
		body, _ := ioutil.ReadAll(res.Body)
		return nil, fmt.Errorf("unexpected HTTP status %d on GET from %s: %s", res.StatusCode, u, string(body))
	}

	resp := &Response{
		Body:          res.Body,
		ContentLength: res.ContentLength,
	}

	hint := ""
	ctypeHdr := res.Header.Get("Content-Type")
	if ctypeHdr != "" {
		mediatype, _, e := mime.ParseMediaType(ctypeHdr)
		if e != nil {
			return nil, e
		}
		hint = mediatype
	}

	resp.ContentType, err = mimeType(u, hint)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
