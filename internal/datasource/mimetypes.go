package datasource

import "mime"

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
