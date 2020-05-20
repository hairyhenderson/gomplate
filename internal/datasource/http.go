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

func (h *HTTP) Read(ctx context.Context, url *url.URL, args ...string) (data *Data, err error) {
	if h.hc == nil {
		h.hc = &http.Client{Timeout: time.Second * 5}
	}

	data = newData(url, args)

	u, err := buildURL(url, args...)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	if hdr, ok := ctx.Value(headerKey).(http.Header); ok {
		req.Header = hdr
	}

	res, err := h.hc.Do(req)
	if err != nil {
		return nil, err
	}
	data.Bytes, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	err = res.Body.Close()
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		err := errors.Errorf("Unexpected HTTP status %d on GET from %s: %s", res.StatusCode, url, string(data.Bytes))
		return nil, err
	}
	ctypeHdr := res.Header.Get("Content-Type")
	if ctypeHdr != "" {
		mediatype, _, e := mime.ParseMediaType(ctypeHdr)
		if e != nil {
			return nil, e
		}
		data.MType = mediatype
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
