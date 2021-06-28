package datasources

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/hairyhenderson/gomplate/v3/internal/datafuncs"
)

// Response - the response to a Request.
type Response struct {
	// Body represents the response body, which may be streamed on-demand if the
	// underlying transport is network-based. It is the caller's responsibility
	// to close Body.
	Body io.ReadCloser
	// ContentType records Body's media type, typically from the Content-Type
	// header or similar mechanisms. May be empty if no specific type was
	// inferred.
	ContentType string
	// ContentLength records Body's length. The value -1
	// indicates that the length is unknown.
	ContentLength int64
}

// Parse - read and parse the response body according to its content type.
func (resp *Response) Parse() (out interface{}, err error) {
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	s := string(b)
	switch mimeAlias(resp.ContentType) {
	case jsonMimetype:
		out, err = datafuncs.JSON(s)
		if err != nil {
			// maybe it's a JSON array
			out, err = datafuncs.JSONArray(s)
		}
	case jsonArrayMimetype:
		out, err = datafuncs.JSONArray(s)
	case yamlMimetype:
		out, err = datafuncs.YAML(s)
		if err != nil {
			// maybe it's a YAML array
			out, err = datafuncs.YAMLArray(s)
		}
	case csvMimetype:
		out, err = datafuncs.CSV(s)
	case tomlMimetype:
		out, err = datafuncs.TOML(s)
	case envMimetype:
		out, err = datafuncs.DotEnv(s)
	case textMimetype:
		out = s
	default:
		return nil, fmt.Errorf("data with content type %s not yet supported", resp.ContentType)
	}
	return out, err
}
