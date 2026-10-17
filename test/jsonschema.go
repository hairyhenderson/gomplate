package test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// schemaURL identifies the schema passed to JSONSchema. It is never loaded;
// it only gives the compiler an absolute location to resolve local $refs
// against.
const schemaURL = "gomplate:///schema.json"

// JSONSchema validates value against the given JSON Schema. The schema can be
// a JSON string, or already-parsed data (for example from a datasource).
// Schemas without a $schema keyword are treated as draft 2020-12. Only
// references within the schema itself are supported - $refs to other files or
// URLs are rejected.
func JSONSchema(schema, value any) error {
	doc, err := schemaDoc(schema)
	if err != nil {
		return fmt.Errorf("invalid schema: %w", err)
	}

	c := jsonschema.NewCompiler()
	c.DefaultDraft(jsonschema.Draft2020)
	c.UseLoader(noLoader{})

	if err = c.AddResource(schemaURL, doc); err != nil {
		return fmt.Errorf("invalid schema: %w", err)
	}

	sch, err := c.Compile(schemaURL)
	if err != nil {
		return fmt.Errorf("invalid schema: %w", err)
	}

	inst, err := toJSONValue(value)
	if err != nil {
		return fmt.Errorf("value can not be validated: %w", err)
	}

	if err := sch.Validate(inst); err != nil {
		return validationError(err)
	}

	return nil
}

// validationError drops the validator's first line, which only names the
// internal schema URL, and keeps the list of failures.
func validationError(err error) error {
	var verr *jsonschema.ValidationError
	if !errors.As(err, &verr) {
		return fmt.Errorf("schema validation failed: %w", err)
	}

	msg := verr.Error()
	if _, causes, ok := strings.Cut(msg, "\n"); ok {
		msg = causes
	}

	return fmt.Errorf("schema validation failed:\n%s", msg)
}

// schemaDoc parses a schema given as a JSON string, or normalises
// already-parsed data.
func schemaDoc(schema any) (any, error) {
	switch s := schema.(type) {
	case string:
		return jsonschema.UnmarshalJSON(strings.NewReader(s))
	case []byte:
		return jsonschema.UnmarshalJSON(bytes.NewReader(s))
	default:
		return toJSONValue(schema)
	}
}

// toJSONValue round-trips a value through JSON, so that the types match what
// the validator expects (e.g. YAML integers become JSON numbers).
func toJSONValue(v any) (any, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return jsonschema.UnmarshalJSON(bytes.NewReader(b))
}

// noLoader refuses to load external schemas, so that validating doesn't
// implicitly read files or make network requests.
type noLoader struct{}

func (noLoader) Load(url string) (any, error) {
	return nil, fmt.Errorf("loading external schemas is not supported: %s", url)
}
