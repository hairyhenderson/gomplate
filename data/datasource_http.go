package data

import (
	"context"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"
)

func buildURL(base *url.URL, args ...string) (*url.URL, error) {
	if len(args) == 0 {
		return base, nil
	}
	p, err := url.Parse(args[0])
	if err != nil {
		return nil, errors.Wrapf(err, "bad sub-path %s", args[0])
	}
	return base.ResolveReference(p), nil
}

func readHTTP(ctx context.Context, source *Source, args ...string) ([]byte, error) {
	if source.hc == nil {
		source.hc = &http.Client{Timeout: time.Second * 5}
	}
	u, err := buildURL(source.URL, args...)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header = source.header
	res, err := source.hc.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	err = res.Body.Close()
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		err := errors.Errorf("Unexpected HTTP status %d on GET from %s: %s", res.StatusCode, source.URL, string(body))
		return nil, err
	}
	ctypeHdr := res.Header.Get("Content-Type")
	if ctypeHdr != "" {
		mediatype, _, e := mime.ParseMediaType(ctypeHdr)
		if e != nil {
			return nil, e
		}
		source.mediaType = mediatype
	}
	return body, nil
}
