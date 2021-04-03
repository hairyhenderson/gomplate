package datasources

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/hairyhenderson/gomplate/v3/coll"
	"github.com/hairyhenderson/gomplate/v3/internal/config"
	"github.com/hairyhenderson/gomplate/v3/internal/datafuncs"
)

// mergeRequester demultiplexes a `merge:` datasource. The 'args' parameter currently
// has no meaning for this source.
//
// URI format is 'merge:<source 1>|<source 2>[|<source n>...]' where `<source #>`
// is a supported URI or a pre-defined alias name.
//
// Query strings and fragments are interpreted relative to the merged data, not
// the source data. To merge datasources with query strings or fragments, define
// separate sources first and specify the alias names. HTTP headers are also not
// supported directly.
type mergeRequester struct {
	ds map[string]config.DataSource
}

func (r *mergeRequester) Request(ctx context.Context, u *url.URL, header http.Header) (*Response, error) {
	opaque := u.Opaque
	parts := strings.Split(opaque, "|")
	if len(parts) < 2 {
		return nil, fmt.Errorf("need at least 2 datasources to merge")
	}

	data := make([]map[string]interface{}, len(parts))
	for i, part := range parts {
		// supports either URIs or aliases
		subSource, err := r.lookup(ctx, part)
		if err != nil {
			// maybe it's a relative filename?
			u, uerr := config.ParseSourceURL(part)
			if uerr != nil {
				return nil, fmt.Errorf("failed to lookup datasource %s: %v, %w", part, err, uerr)
			}
			subSource = config.DataSource{URL: u}
		}

		resp, err := Request(ctx, subSource)
		if err != nil {
			return nil, fmt.Errorf("couldn't read datasource %q: %w", part, err)
		}

		data[i], err = parseMap(resp)
		if err != nil {
			return nil, err
		}
	}

	// Merge the data together
	b, err := mergeData(data)
	if err != nil {
		return nil, err
	}

	resp := &Response{
		ContentType: yamlMimetype,
		Body:        ioutil.NopCloser(bytes.NewReader(b)),
	}

	return resp, nil
}

func (r *mergeRequester) lookup(ctx context.Context, alias string) (config.DataSource, error) {
	d := config.DataSourcesFromContext(ctx)
	for k, v := range d {
		r.ds[k] = v
	}

	ds, ok := r.ds[alias]
	if !ok {
		srcURL, err := url.Parse(alias)
		if err != nil || !srcURL.IsAbs() {
			return ds, fmt.Errorf("undefined datasource %q", alias)
		}

		ds.URL = srcURL
		// we can't support HTTP headers directly...
		// ds.Header = d.extraHeaders[alias]

		r.ds[alias] = ds
	}

	return ds, nil
}

func mergeData(data []map[string]interface{}) (out []byte, err error) {
	dst := data[0]
	data = data[1:]

	dst, err = coll.Merge(dst, data...)
	if err != nil {
		return nil, err
	}

	s, err := datafuncs.ToYAML(dst)
	if err != nil {
		return nil, err
	}

	return []byte(s), nil
}

func parseMap(resp *Response) (map[string]interface{}, error) {
	datum, err := resp.Parse()
	if err != nil {
		return nil, err
	}

	var m map[string]interface{}
	switch datum := datum.(type) {
	case map[string]interface{}:
		m = datum
	default:
		return nil, fmt.Errorf("unexpected data type '%T' for datasource (type %s); merge: can only merge maps", datum, resp.ContentType)
	}

	return m, nil
}
