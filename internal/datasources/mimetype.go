package datasources

import (
	"fmt"
	"mime"
	"net/url"
	"path/filepath"
	"strings"
)

// mimeType returns the MIME type to use to parse some response.
//
// The MIME type is determined by these rules:
// 1. the 'type' URL query parameter is used if present
// 2. otherwise, the given hint is used, if present
// 3. otherwise, a MIME type is calculated from the file extension, if the extension is registered
// 4. otherwise, the default type of 'text/plain' is used
func mimeType(u *url.URL, hint string) (mimeType string, err error) {
	mediatype := u.Query().Get("type")

	if mediatype == "" {
		mediatype = hint
	}

	// make it so + doesn't need to be escaped
	mediatype = strings.ReplaceAll(mediatype, " ", "+")

	if mediatype == "" {
		ext := filepath.Ext(u.Path)
		mediatype = mime.TypeByExtension(ext)
	}

	if mediatype != "" {
		t, _, err := mime.ParseMediaType(mediatype)
		if err != nil {
			return "", fmt.Errorf("failed to parse Media Type %q: %w", mediatype, err)
		}
		mediatype = t
		return mediatype, nil
	}

	return textMimetype, nil
}

func regExtension(ext, typ string) {
	err := mime.AddExtensionType(ext, typ)
	if err != nil {
		panic(err)
	}
}

func init() {
	// Add some types we want to be able to handle which can be missing by default
	regExtension(".json", jsonMimetype)
	regExtension(".yml", yamlMimetype)
	regExtension(".yaml", yamlMimetype)
	regExtension(".csv", csvMimetype)
	regExtension(".toml", tomlMimetype)
	regExtension(".env", envMimetype)
}

const (
	textMimetype      = "text/plain"
	csvMimetype       = "text/csv"
	jsonMimetype      = "application/json"
	jsonArrayMimetype = "application/array+json"
	tomlMimetype      = "application/toml"
	yamlMimetype      = "application/yaml"
	envMimetype       = "application/x-env"
)

// mimeTypeAliases defines a mapping for non-canonical mime types that are
// sometimes seen in the wild
var mimeTypeAliases = map[string]string{
	"application/x-yaml": yamlMimetype,
	"application/text":   textMimetype,
}

func mimeAlias(m string) string {
	if a, ok := mimeTypeAliases[m]; ok {
		return a
	}
	return m
}
