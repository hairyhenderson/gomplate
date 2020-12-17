package datasources

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/hairyhenderson/gomplate/v3/env"
)

type envRequester struct{}

func (r *envRequester) Request(ctx context.Context, u *url.URL, header http.Header) (*Response, error) {
	n := u.Path
	n = strings.TrimPrefix(n, "/")
	if n == "" {
		n = u.Opaque
	}

	s := env.Getenv(n)

	ct, err := mimeType(u, textMimetype)
	if err != nil {
		ct = textMimetype
	}
	resp := &Response{
		Body:          ioutil.NopCloser(bytes.NewBufferString(s)),
		ContentLength: int64(len(s)),
		ContentType:   ct,
	}
	return resp, nil
}
