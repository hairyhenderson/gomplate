package parsers

import (
	"fmt"

	"github.com/hairyhenderson/gomplate/v4/internal/iohelpers"
)

func ParseData(mimeType, s string) (out any, err error) {
	switch iohelpers.MimeAlias(mimeType) {
	case iohelpers.JSONMimetype:
		out, err = JSON(s)
		if err != nil {
			// maybe it's a JSON array
			out, err = JSONArray(s)
		}
	case iohelpers.JSONArrayMimetype:
		out, err = JSONArray(s)
	case iohelpers.YAMLMimetype:
		out, err = YAML(s)
		if err != nil {
			// maybe it's a YAML array
			out, err = YAMLArray(s)
		}
	case iohelpers.CSVMimetype:
		out, err = CSV(s)
	case iohelpers.TOMLMimetype:
		out, err = TOML(s)
	case iohelpers.EnvMimetype:
		out, err = DotEnv(s)
	case iohelpers.TextMimetype:
		out = s
	case iohelpers.CUEMimetype:
		out, err = CUE(s)
	default:
		return nil, fmt.Errorf("data of type %q not yet supported", mimeType)
	}
	return out, err
}
