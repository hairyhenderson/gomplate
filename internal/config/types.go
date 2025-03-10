package config

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/hairyhenderson/gomplate/v4/internal/deprecated"
	"github.com/hairyhenderson/gomplate/v4/internal/urlhelpers"
	"github.com/hairyhenderson/yaml"
)

// Templates - a map of templates. We can't just use map[string]DataSource,
// because we need to be able to marshal both the old (array of '[k=]v' strings)
// and the new (proper map) formats.
//
// Note that templates use the DataSource type, since they have the exact same
// shape.
// TODO: get rid of this and just use map[string]DataSource once the legacy
// [k=]v array format is no longer supported (v4.1.0?)
type Templates map[string]DataSource

// UnmarshalYAML - satisfy the yaml.Umarshaler interface
func (t *Templates) UnmarshalYAML(value *yaml.Node) error {
	// first attempt to unmarshal as a map[string]DataSource
	m := map[string]DataSource{}
	err := value.Decode(m)
	if err == nil {
		*t = m
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
	deprecated.WarnDeprecated(context.Background(),
		"config: the YAML array form for 'templates' is deprecated and will be removed in the next version. Use the map form instead.")
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

		u, err := urlhelpers.ParseSourceURL(pth)
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

func (t Templates) MarshalYAML() (any, error) {
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

type experimentalCtxKey struct{}

func SetExperimental(ctx context.Context) context.Context {
	return context.WithValue(ctx, experimentalCtxKey{}, true)
}

func ExperimentalEnabled(ctx context.Context) bool {
	v, ok := ctx.Value(experimentalCtxKey{}).(bool)
	return ok && v
}

// DataSource - datasource configuration
//
// defined in this package to avoid cyclic dependencies
type DataSource struct {
	URL    *url.URL    `yaml:"-"`
	Header http.Header `yaml:"header,omitempty,flow"`
}

// UnmarshalYAML - satisfy the yaml.Umarshaler interface - URLs aren't
// well supported, and anyway we need to do some extra parsing
func (d *DataSource) UnmarshalYAML(value *yaml.Node) error {
	type raw struct {
		Header http.Header
		URL    string
	}
	r := raw{}
	err := value.Decode(&r)
	if err != nil {
		return err
	}
	u, err := urlhelpers.ParseSourceURL(r.URL)
	if err != nil {
		return fmt.Errorf("could not parse datasource URL %q: %w", r.URL, err)
	}
	*d = DataSource{
		URL:    u,
		Header: r.Header,
	}
	return nil
}

// MarshalYAML - satisfy the yaml.Marshaler interface - URLs aren't
// well supported, and anyway we need to do some extra parsing
func (d DataSource) MarshalYAML() (any, error) {
	type raw struct {
		Header http.Header
		URL    string
	}
	r := raw{
		URL:    d.URL.String(),
		Header: d.Header,
	}
	return r, nil
}
