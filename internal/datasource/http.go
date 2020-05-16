package datasource

import (
	"context"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"
)

// HTTP -
type HTTP struct {
	hc *http.Client
}

var _ Reader = (*HTTP)(nil)

func (h *HTTP) Read(ctx context.Context, url *url.URL, args ...string) (data Data, err error) {
	if h.hc == nil {
		h.hc = &http.Client{Timeout: time.Second * 5}
	}
	u, err := buildURL(url, args...)
	if err != nil {
		return data, err
	}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return data, err
	}
	// TODO: uncomment
	// req.Header = source.header

	res, err := h.hc.Do(req)
	if err != nil {
		return data, err
	}
	data.Bytes, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return data, err
	}
	err = res.Body.Close()
	if err != nil {
		return data, err
	}
	if res.StatusCode != 200 {
		err := errors.Errorf("Unexpected HTTP status %d on GET from %s: %s", res.StatusCode, url, string(data.Bytes))
		return data, err
	}
	ctypeHdr := res.Header.Get("Content-Type")
	if ctypeHdr != "" {
		mediatype, _, e := mime.ParseMediaType(ctypeHdr)
		if e != nil {
			return data, e
		}
		data.MediaType = mediatype
	}
	return data, nil
}

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
