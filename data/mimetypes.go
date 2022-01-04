package data

import (
	"mime"
)

const (
	textMimetype      = "text/plain"
	csvMimetype       = "text/csv"
	jsonMimetype      = "application/json"
	jsonArrayMimetype = "application/array+json"
	tomlMimetype      = "application/toml"
	yamlMimetype      = "application/yaml"
	envMimetype       = "application/x-env"
	cueMimetype       = "application/cue"
)

// mimeTypeAliases defines a mapping for non-canonical mime types that are
// sometimes seen in the wild
var mimeTypeAliases = map[string]string{
	"application/x-yaml": yamlMimetype,
	"application/text":   textMimetype,
}

func mimeAlias(m string) string {
	// normalize the type by removing any extra parameters
	m, _, _ = mime.ParseMediaType(m)

	if a, ok := mimeTypeAliases[m]; ok {
		return a
	}
	return m
}
