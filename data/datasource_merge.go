package data

import (
	"context"
	"strings"

	"github.com/hairyhenderson/gomplate/v3/coll"
	"github.com/hairyhenderson/gomplate/v3/internal/config"

	"github.com/pkg/errors"
)

// readMerge demultiplexes a `merge:` datasource. The 'args' parameter currently
// has no meaning for this source.
//
// URI format is 'merge:<source 1>|<source 2>[|<source n>...]' where `<source #>`
// is a supported URI or a pre-defined alias name.
//
// Query strings and fragments are interpreted relative to the merged data, not
// the source data. To merge datasources with query strings or fragments, define
// separate sources first and specify the alias names. HTTP headers are also not
// supported directly.
func (d *Data) readMerge(ctx context.Context, source *Source, _ ...string) ([]byte, error) {
	opaque := source.URL.Opaque
	parts := strings.Split(opaque, "|")
	if len(parts) < 2 {
		return nil, errors.New("need at least 2 datasources to merge")
	}
	data := make([]map[string]interface{}, len(parts))
	for i, part := range parts {
		// supports either URIs or aliases
		subSource, err := d.lookupSource(part)
		if err != nil {
			// maybe it's a relative filename?
			u, uerr := config.ParseSourceURL(part)
			if uerr != nil {
				return nil, uerr
			}
			subSource = &Source{
				Alias: part,
				URL:   u,
			}
		}
		subSource.inherit(source)

		b, err := d.readSource(ctx, subSource)
		if err != nil {
			return nil, errors.Wrapf(err, "Couldn't read datasource '%s'", part)
		}

		mimeType, err := subSource.mimeType("")
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read datasource %s", subSource.URL)
		}

		data[i], err = parseMap(mimeType, string(b))
		if err != nil {
			return nil, err
		}
	}

	// Merge the data together
	b, err := mergeData(data)
	if err != nil {
		return nil, err
	}

	source.mediaType = yamlMimetype
	return b, nil
}

func mergeData(data []map[string]interface{}) (out []byte, err error) {
	dst := data[0]
	data = data[1:]

	dst, err = coll.Merge(dst, data...)
	if err != nil {
		return nil, err
	}

	s, err := ToYAML(dst)
	if err != nil {
		return nil, err
	}
	return []byte(s), nil
}

func parseMap(mimeType, data string) (map[string]interface{}, error) {
	datum, err := parseData(mimeType, data)
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	switch datum := datum.(type) {
	case map[string]interface{}:
		m = datum
	default:
		return nil, errors.Errorf("unexpected data type '%T' for datasource (type %s); merge: can only merge maps", datum, mimeType)
	}
	return m, nil
}
