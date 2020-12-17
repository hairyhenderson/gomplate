package datasources

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

type stdinRequester struct {
	stdin io.Reader
}

func (r *stdinRequester) Request(ctx context.Context, u *url.URL, header http.Header) (resp *Response, err error) {
	if r.stdin == nil {
		r.stdin = os.Stdin
	}

	resp = &Response{
		// Even though os.Stdin is actually an io.ReadCloser, wrap with
		// NopCloser. Closing os.Stdin has undesired side-effects.
		Body:          ioutil.NopCloser(r.stdin),
		ContentLength: -1,
	}

	resp.ContentType, err = mimeType(u, "")
	if err != nil {
		return nil, err
	}

	return resp, nil
}
