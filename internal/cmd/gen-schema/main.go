// gen-schema generates a JSON Schema for the .gomplate.yaml configuration file.
// Run via: make schema/gomplate-config.json
package main

import (
	"encoding/json"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"reflect"
	"runtime/debug"
	"time"

	"github.com/hairyhenderson/gomplate/v5"
	"github.com/hairyhenderson/gomplate/v5/internal/config"
	"github.com/invopop/jsonschema"
)

var (
	typeDataSource   = reflect.TypeFor[config.DataSource]()
	typePluginConfig = reflect.TypeFor[gomplate.PluginConfig]()
	typeDuration     = reflect.TypeFor[time.Duration]()
	typeHTTPHeader   = reflect.TypeFor[http.Header]()
)

func main() {
	out := flag.String("out", "schema/gomplate-config.json", "output file path")
	flag.Parse()

	r := &jsonschema.Reflector{
		// Use yaml struct tags for field names instead of json tags
		FieldNameTag: "yaml",
		// Only mark fields as required when they have jsonschema:required tag
		RequiredFromJSONSchemaTags: true,
		// Inline the root type rather than wrapping it in a $ref
		ExpandedStruct: true,
		// Mapper handles types with custom YAML unmarshalling or poor reflection
		Mapper: typeMapper,
	}

	modPath := modulePath()
	if err := r.AddGoComments(modPath, "."); err != nil {
		slog.Warn("could not add Go comments", "err", err)
	}
	if err := r.AddGoComments(modPath, "./internal/config"); err != nil {
		slog.Warn("could not add Go comments", "err", err)
	}

	schema := r.Reflect(&gomplate.Config{})
	schema.ID = "https://raw.githubusercontent.com/hairyhenderson/gomplate/main/schema/gomplate-config.json"
	schema.Title = "gomplate configuration"
	schema.Description = "Configuration file for gomplate (.gomplate.yaml / .gomplate.yml)"

	// Add constraints that reflection can't derive automatically.
	if p, ok := schema.Properties.Get("missingKey"); ok {
		p.Enum = []any{"", "error", "zero", "default", "invalid"}
	}
	if p, ok := schema.Properties.Get("pluginTimeout"); ok {
		p.Description = "timeout for all plugins, e.g. 500ms, 5s (default 5s)"
	}

	b, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		slog.Error("error marshalling schema", "err", err)
		os.Exit(1)
	}
	b = append(b, '\n')

	if err := os.WriteFile(*out, b, 0o644); err != nil { //nolint:gosec // schema file is intended to be world-readable
		slog.Error("error writing schema", "err", err)
		os.Exit(1)
	}
}

func typeMapper(t reflect.Type) *jsonschema.Schema {
	switch t {
	case typeDataSource:
		return dataSourceSchema()
	case typePluginConfig:
		return pluginConfigSchema()
	case typeDuration:
		return &jsonschema.Schema{
			Type:        "string",
			Description: "duration string, e.g. 30s, 5m, 1h",
		}
	case typeHTTPHeader:
		return httpHeaderSchema()
	}
	return nil
}

func dataSourceSchema() *jsonschema.Schema {
	props := jsonschema.NewProperties()
	props.Set("url", &jsonschema.Schema{
		Type:        "string",
		Description: "URL for the datasource (e.g. file:///data.json, https://example.com/data, env:FOO)",
	})
	props.Set("header", httpHeaderSchema())
	return &jsonschema.Schema{
		Type:                 "object",
		Description:          "Data source configuration",
		Properties:           props,
		AdditionalProperties: jsonschema.FalseSchema,
	}
}

func pluginConfigSchema() *jsonschema.Schema {
	props := jsonschema.NewProperties()
	props.Set("cmd", &jsonschema.Schema{
		Type:        "string",
		Description: "Path to the plugin command",
	})
	props.Set("args", &jsonschema.Schema{
		Type:  "array",
		Items: &jsonschema.Schema{Type: "string"},
	})
	props.Set("timeout", &jsonschema.Schema{
		Type:        "string",
		Description: "Plugin execution timeout, e.g. 30s, 5m",
	})
	props.Set("pipe", &jsonschema.Schema{
		Type:        "boolean",
		Description: "Pipe gomplate's stdin to the plugin",
	})
	return &jsonschema.Schema{
		OneOf: []*jsonschema.Schema{
			{Type: "string", Description: "Path to the plugin command"},
			{
				Type:                 "object",
				Description:          "Plugin configuration",
				Properties:           props,
				Required:             []string{"cmd"},
				AdditionalProperties: jsonschema.FalseSchema,
			},
		},
	}
}

func modulePath() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		return info.Main.Path
	}
	return "github.com/hairyhenderson/gomplate/v5"
}

func httpHeaderSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:        "object",
		Description: "HTTP headers (header name → list of values)",
		AdditionalProperties: &jsonschema.Schema{
			Type:  "array",
			Items: &jsonschema.Schema{Type: "string"},
		},
	}
}
