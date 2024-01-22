package iohelpers

import (
	"mime"
)

const (
	TextMimetype      = "text/plain"
	CSVMimetype       = "text/csv"
	JSONMimetype      = "application/json"
	JSONArrayMimetype = "application/array+json"
	TOMLMimetype      = "application/toml"
	YAMLMimetype      = "application/yaml"
	EnvMimetype       = "application/x-env"
	CUEMimetype       = "application/cue"
)

// mimeTypeAliases defines a mapping for non-canonical mime types that are
// sometimes seen in the wild
var mimeTypeAliases = map[string]string{
	"application/x-yaml": YAMLMimetype,
	"application/text":   TextMimetype,
}

func MimeAlias(m string) string {
	// normalize the type by removing any extra parameters
	m, _, _ = mime.ParseMediaType(m)

	if a, ok := mimeTypeAliases[m]; ok {
		return a
	}
	return m
}
