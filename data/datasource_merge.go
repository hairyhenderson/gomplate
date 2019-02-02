package data

import (
	"strings"

	"github.com/imdario/mergo"

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
func (d *Data) readMerge(source *Source, args ...string) ([]byte, error) {
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
			subSource, err = parseSource(part + "=" + part)
			if err != nil {
				return nil, err
			}
		}
		subSource.inherit(source)

		b, err := d.readSource(subSource)
		if err != nil {
			return nil, errors.Wrapf(err, "Couldn't read datasource '%s'", part)
		}

		mimeType, err := subSource.mimeType()
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

func mergeData(data []map[string]interface{}) ([]byte, error) {
	dst := data[0]
	data = data[1:]
	for _, datum := range data {
		err := mergo.Merge(&dst, datum)
		if err != nil {
			return nil, errors.Wrap(err, "failed to merge datasources")
		}
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
