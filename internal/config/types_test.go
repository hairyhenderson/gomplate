package config

import (
	"net/http"
	"testing"

	"github.com/hairyhenderson/yaml"
	"github.com/stretchr/testify/assert"
)

func TestTemplates_UnmarshalYAML(t *testing.T) {
	in := `t:
  url: foo/bar/helloworld.tmpl
templatedir:
  url: templatedir/
dir:
  url: foo/bar/
mytemplate.t:
  url: mytemplate.t
remote:
  url: https://example.com/foo/bar/helloworld.tmpl
  header:
    Accept: [text/plain, text/template]`
	out := Templates{}
	err := yaml.Unmarshal([]byte(in), &out)
	assert.NoError(t, err)
	assert.EqualValues(t, Templates{
		"t":            {URL: mustURL("foo/bar/helloworld.tmpl")},
		"templatedir":  {URL: mustURL("templatedir/")},
		"dir":          {URL: mustURL("foo/bar/")},
		"mytemplate.t": {URL: mustURL("mytemplate.t")},
		"remote": {
			URL:    mustURL("https://example.com/foo/bar/helloworld.tmpl"),
			Header: http.Header{"Accept": {"text/plain", "text/template"}},
		},
	}, out)

	// legacy array format
	in = `- t=foo/bar/helloworld.tmpl
- templatedir/
- dir=foo/bar/
- mytemplate.t
- remote=https://example.com/foo/bar/helloworld.tmpl`
	out = Templates{}
	err = yaml.Unmarshal([]byte(in), &out)
	assert.NoError(t, err)
	assert.EqualValues(t, Templates{
		"t":            {URL: mustURL("foo/bar/helloworld.tmpl")},
		"templatedir/": {URL: mustURL("templatedir/")},
		"dir":          {URL: mustURL("foo/bar/")},
		"mytemplate.t": {URL: mustURL("mytemplate.t")},
		"remote":       {URL: mustURL("https://example.com/foo/bar/helloworld.tmpl")},
	}, out)

	// invalid format
	in = `"neither an array nor a map"`
	out = Templates{}
	err = yaml.Unmarshal([]byte(in), &out)
	assert.Error(t, err)

	// invalid URL
	in = `- t="not a:valid url"`
	out = Templates{}
	err = yaml.Unmarshal([]byte(in), &out)
	assert.Error(t, err)
}

func TestParseTemplateArg(t *testing.T) {
	data := []struct {
		ds    DataSource
		in    string
		alias string
	}{
		{in: "t=foo/bar/helloworld.tmpl", alias: "t", ds: DataSource{URL: mustURL("foo/bar/helloworld.tmpl")}},
		{in: "templatedir/", alias: "templatedir/", ds: DataSource{URL: mustURL("templatedir/")}},
		{in: "dir=foo/bar/", alias: "dir", ds: DataSource{URL: mustURL("foo/bar/")}},
		{in: "mytemplate.t", alias: "mytemplate.t", ds: DataSource{URL: mustURL("mytemplate.t")}},
		{
			in:    "remote=https://example.com/foo/bar/helloworld.tmpl",
			alias: "remote", ds: DataSource{URL: mustURL("https://example.com/foo/bar/helloworld.tmpl")},
		},
	}

	for _, d := range data {
		alias, ds, err := parseTemplateArg(d.in)
		assert.NoError(t, err)
		assert.Equal(t, d.alias, alias)
		assert.EqualValues(t, d.ds, ds)
	}
}
