package config

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/hairyhenderson/gomplate/v5/internal/urlhelpers"
	"github.com/hairyhenderson/yaml"
)

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
