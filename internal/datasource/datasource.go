package datasource

import (
	"context"
	"net/url"
)

// Reader -
type Reader interface {
	Read(ctx context.Context, url *url.URL, args ...string) (data Data, err error)
}

// Data - some data read from a Reader.
type Data struct {
	Bytes     []byte
	MediaType string
}
