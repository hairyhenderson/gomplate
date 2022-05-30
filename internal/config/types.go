package config

import (
	"fmt"
	"net/http"
	"strings"

	"gopkg.in/yaml.v3"
)

// Templates - a map of templates. We can't just use map[string]DataSource,
// because we need to be able to marshal both the old (array of '[k=]v' strings)
// and the new (proper map) formats.
//
// Note that templates use the DataSource type, since they have the exact same
// shape.
// TODO: get rid of this and just use map[string]DataSource once the legacy
// [k=]v array format is no longer supported
type Templates map[string]DataSource

// UnmarshalYAML - satisfy the yaml.Umarshaler interface
func (t *Templates) UnmarshalYAML(value *yaml.Node) error {
	// first attempt to unmarshal as a map[string]DataSource
	err := value.Decode(map[string]DataSource(*t))
	if err == nil {
		return nil
	}

	// if that fails, try to unmarshal as an array of '[k=]v' strings
	err = t.unmarshalYAMLArray(value)
	if err != nil {
		return fmt.Errorf("could not unmarshal templates as map or array: %w", err)
	}

	return nil
}

func (t *Templates) unmarshalYAMLArray(value *yaml.Node) error {
	a := []string{}
	err := value.Decode(&a)
	if err != nil {
		return fmt.Errorf("could not unmarshal templates as array: %w", err)
	}

	ts := Templates{}
	for _, s := range a {
		alias, pth, _ := strings.Cut(s, "=")
		if pth == "" {
			// when alias is omitted, the path and alias are identical
			pth = alias
		}

		u, err := ParseSourceURL(pth)
		if err != nil {
			return fmt.Errorf("could not parse template URL %q: %w", pth, err)
		}

		ts[alias] = DataSource{
			URL: u,
		}
	}

	*t = ts

	return nil
}

func (t Templates) MarshalYAML() (interface{}, error) {
	type rawTemplate struct {
		Header http.Header `yaml:"header,omitempty,flow"`
		URL    string      `yaml:"url"`
	}

	m := map[string]rawTemplate{}
	for k, v := range t {
		m[k] = rawTemplate{
			Header: v.Header,
			URL:    v.URL.String(),
		}
	}
	return m, nil
}

func parseTemplateArg(value string) (alias string, ds DataSource, err error) {
	alias, u, _ := strings.Cut(value, "=")
	if u == "" {
		u = alias
	}

	ds.URL, err = ParseSourceURL(u)

	return alias, ds, err
}
