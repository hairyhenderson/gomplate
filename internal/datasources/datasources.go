package datasources

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/hairyhenderson/gomplate/v3/internal/config"
)

// requester is the interface that wraps the Request method. Implementations
// will not typically support many different URL schemes.
type requester interface {
	// Request some data from a datasource
	Request(ctx context.Context, url *url.URL, header http.Header) (*Response, error)
}

type urlBuilder interface {
	BuildURL(url *url.URL, args ...string) (*url.URL, error)
}

// ReadDataSource reads the (potentially cached) data from the given DataSource,
// as referenced by the given args.
func ReadDataSource(ctx context.Context, ds config.DataSource, args ...string) (string, []byte, error) {
	resp, err := Request(ctx, ds, args...)
	if err != nil {
		return "", nil, err
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)

	return resp.ContentType, data, err
}

// Request requests the (potentially cached) data from the given DataSource,
// as referenced by the given args.
func Request(ctx context.Context, ds config.DataSource, args ...string) (*Response, error) {
	// TODO: caching

	if ds.URL == nil {
		// TODO: support dynamic datasources
		return nil, fmt.Errorf("invalid datasource, nil URL: %#v", ds)
	}

	ub, err := lookupURLBuilder(ds.URL.Scheme)
	if err != nil {
		return nil, fmt.Errorf("failed to find urlBuilder for %v: %w", ds, err)
	}

	u, err := ub.BuildURL(ds.URL, args...)
	if err != nil {
		return nil, err
	}

	rdr, err := lookupRequester(ctx, u.Scheme)
	if err != nil {
		return nil, fmt.Errorf("failed to find requester for %s: %w", u, err)
	}

	resp, err := rdr.Request(ctx, u, ds.Header)
	if err != nil {
		return nil, fmt.Errorf("failed to request %s: %w", u, err)
	}
	return resp, err
}
