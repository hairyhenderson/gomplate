package test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const personSchema = `{
	"type": "object",
	"required": ["name"],
	"additionalProperties": false,
	"properties": {
		"name": {"type": "string"},
		"age": {"$ref": "#/$defs/age"}
	},
	"$defs": {
		"age": {"type": "integer", "minimum": 0}
	}
}`

func TestJSONSchema(t *testing.T) {
	t.Parallel()

	// schema as a JSON string, value as parsed data with a Go int
	require.NoError(t, JSONSchema(personSchema, map[string]any{"name": "Jane", "age": 42}))
	require.NoError(t, JSONSchema([]byte(personSchema), map[string]any{"name": "Jane"}))

	// schema as parsed data
	schema := map[string]any{
		"type":  "array",
		"items": map[string]any{"enum": []any{"a", "b"}},
	}
	require.NoError(t, JSONSchema(schema, []any{"a", "b"}))
	require.ErrorContains(t, JSONSchema(schema, []any{"a", "c"}), "schema validation failed")

	// a plain string value is validated as a string, not parsed as JSON
	require.NoError(t, JSONSchema(`{"type": "string"}`, `{"not": "an object"}`))

	// draft is taken from $schema when present
	require.NoError(t, JSONSchema(`{"$schema": "http://json-schema.org/draft-07/schema#", "type": "string"}`, "x"))
}

func TestJSONSchema_Invalid(t *testing.T) {
	t.Parallel()

	testdata := []struct {
		value any
		msg   string
	}{
		{map[string]any{"age": 42}, "missing property 'name'"},
		{map[string]any{"name": "Jane", "age": -1}, "/age"},
		{map[string]any{"name": "Jane", "nickname": "JJ"}, "additional properties 'nickname' not allowed"},
		{map[string]any{"name": 42}, "/name"},
	}

	for _, d := range testdata {
		err := JSONSchema(personSchema, d.value)
		require.ErrorContains(t, err, "schema validation failed")
		require.ErrorContains(t, err, d.msg)
	}
}

func TestJSONSchema_BadSchema(t *testing.T) {
	t.Parallel()

	require.ErrorContains(t, JSONSchema(`{"type": `, "x"), "invalid schema")
	require.ErrorContains(t, JSONSchema(`{"type": "nonsense"}`, "x"), "invalid schema")

	// external references are never loaded
	err := JSONSchema(`{"$ref": "other.json"}`, "x")
	require.ErrorContains(t, err, "loading external schemas is not supported")
	err = JSONSchema(`{"$ref": "https://example.com/schema.json"}`, "x")
	require.ErrorContains(t, err, "loading external schemas is not supported")
}
