package gomplate

import (
	"bytes"
	"context"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"

	"text/template"

	"github.com/hairyhenderson/gomplate/v3/aws"
	"github.com/hairyhenderson/gomplate/v3/conv"
	"github.com/hairyhenderson/gomplate/v3/data"
	"github.com/hairyhenderson/gomplate/v3/env"
	"github.com/hairyhenderson/gomplate/v3/internal/writers"

	"github.com/stretchr/testify/assert"
)

func testTemplate(g *gomplate, tmpl string) string {
	var out bytes.Buffer
	err := g.runTemplate(context.Background(), &tplate{name: "testtemplate", contents: tmpl, target: &out})
	if err != nil {
		panic(err)
	}
	return out.String()
}

func TestGetenvTemplates(t *testing.T) {
	g := &gomplate{
		funcMap: template.FuncMap{
			"getenv": env.Getenv,
			"bool":   conv.Bool,
		},
	}
	assert.Empty(t, testTemplate(g, `{{getenv "BLAHBLAHBLAH"}}`))
	assert.Equal(t, os.Getenv("USER"), testTemplate(g, `{{getenv "USER"}}`))
	assert.Equal(t, "default value", testTemplate(g, `{{getenv "BLAHBLAHBLAH" "default value"}}`))
}

func TestBoolTemplates(t *testing.T) {
	g := &gomplate{
		funcMap: template.FuncMap{
			"bool": conv.Bool,
		},
	}
	assert.Equal(t, "true", testTemplate(g, `{{bool "true"}}`))
	assert.Equal(t, "false", testTemplate(g, `{{bool "false"}}`))
	assert.Equal(t, "false", testTemplate(g, `{{bool "foo"}}`))
	assert.Equal(t, "false", testTemplate(g, `{{bool ""}}`))
}

func TestEc2MetaTemplates(t *testing.T) {
	createGomplate := func(status int, body string) (*gomplate, *httptest.Server) {
		server, ec2meta := aws.MockServer(status, body)
		return &gomplate{funcMap: template.FuncMap{"ec2meta": ec2meta.Meta}}, server
	}

	g, s := createGomplate(404, "")
	defer s.Close()
	assert.Equal(t, "", testTemplate(g, `{{ec2meta "foo"}}`))
	assert.Equal(t, "default", testTemplate(g, `{{ec2meta "foo" "default"}}`))

	s.Close()
	g, s = createGomplate(200, "i-1234")
	defer s.Close()
	assert.Equal(t, "i-1234", testTemplate(g, `{{ec2meta "instance-id"}}`))
	assert.Equal(t, "i-1234", testTemplate(g, `{{ec2meta "instance-id" "default"}}`))
}

func TestEc2MetaTemplates_WithJSON(t *testing.T) {
	server, ec2meta := aws.MockServer(200, `{"foo":"bar"}`)
	defer server.Close()
	g := &gomplate{
		funcMap: template.FuncMap{
			"ec2meta":    ec2meta.Meta,
			"ec2dynamic": ec2meta.Dynamic,
			"json":       data.JSON,
		},
	}

	assert.Equal(t, "bar", testTemplate(g, `{{ (ec2meta "obj" | json).foo }}`))
	assert.Equal(t, "bar", testTemplate(g, `{{ (ec2dynamic "obj" | json).foo }}`))
}

func TestJSONArrayTemplates(t *testing.T) {
	g := &gomplate{
		funcMap: template.FuncMap{
			"jsonArray": data.JSONArray,
		},
	}

	assert.Equal(t, "[foo bar]", testTemplate(g, `{{jsonArray "[\"foo\",\"bar\"]"}}`))
	assert.Equal(t, "bar", testTemplate(g, `{{ index (jsonArray "[\"foo\",\"bar\"]") 1 }}`))
}

func TestYAMLTemplates(t *testing.T) {
	g := &gomplate{
		funcMap: template.FuncMap{
			"yaml":      data.YAML,
			"yamlArray": data.YAMLArray,
		},
	}

	assert.Equal(t, "bar", testTemplate(g, `{{(yaml "foo: bar").foo}}`))
	assert.Equal(t, "[foo bar]", testTemplate(g, `{{yamlArray "- foo\n- bar\n"}}`))
	assert.Equal(t, "bar", testTemplate(g, `{{ index (yamlArray "[\"foo\",\"bar\"]") 1 }}`))
}

func TestSliceTemplates(t *testing.T) {
	g := &gomplate{
		funcMap: template.FuncMap{
			"slice": conv.Slice,
		},
	}
	assert.Equal(t, "foo", testTemplate(g, `{{index (slice "foo") 0}}`))
	assert.Equal(t, `[foo bar 42]`, testTemplate(g, `{{slice "foo" "bar" 42}}`))
	assert.Equal(t, `helloworld`, testTemplate(g, `{{range slice "hello" "world"}}{{.}}{{end}}`))
}

func TestHasTemplate(t *testing.T) {
	g := &gomplate{
		funcMap: template.FuncMap{
			"yaml": data.YAML,
			"has":  conv.Has,
		},
	}
	assert.Equal(t, "true", testTemplate(g, `{{has ("foo:\n  bar: true" | yaml) "foo"}}`))
	assert.Equal(t, "true", testTemplate(g, `{{has ("foo:\n  bar: true" | yaml).foo "bar"}}`))
	assert.Equal(t, "false", testTemplate(g, `{{has ("foo: true" | yaml) "bah"}}`))
	tmpl := `{{- $data := yaml "foo: bar\nbaz: qux\n" }}
{{- if (has $data "baz") }}
{{- $data.baz }}
{{- end }}`
	assert.Equal(t, "qux", testTemplate(g, tmpl))
	tmpl = `{{- $data := yaml "foo: bar\nbaz: qux\n" }}
{{- if (has $data "quux") }}
{{- $data.quux }}
{{- else }}
{{- $data.foo }}
{{- end }}`
	assert.Equal(t, "bar", testTemplate(g, tmpl))
}

func TestCustomDelim(t *testing.T) {
	g := &gomplate{
		leftDelim:  "[",
		rightDelim: "]",
		funcMap:    template.FuncMap{},
	}
	assert.Equal(t, "hi", testTemplate(g, `[print "hi"]`))
}

func TestRunTemplates(t *testing.T) {
	defer func() { Stdout = os.Stdout }()
	buf := &bytes.Buffer{}
	Stdout = &writers.NopCloser{Writer: buf}
	config := &Config{Input: "foo", OutputFiles: []string{"-"}}
	err := RunTemplates(config)
	assert.NoError(t, err)
	assert.Equal(t, "foo", buf.String())
	assert.Equal(t, 1, Metrics.TemplatesGathered)
	assert.Equal(t, 1, Metrics.TemplatesProcessed)
	assert.Equal(t, 0, Metrics.Errors)
}

func TestParseTemplateArg(t *testing.T) {
	fs = afero.NewMemMapFs()
	afero.WriteFile(fs, "foo.t", []byte("hi"), 0600)
	_ = fs.MkdirAll("dir", 0755)
	afero.WriteFile(fs, "dir/foo.t", []byte("hi"), 0600)
	afero.WriteFile(fs, "dir/bar.t", []byte("hi"), 0600)

	testdata := []struct {
		arg      string
		expected map[string]string
		err      bool
	}{
		{"bogus.t", nil, true},
		{"foo.t", map[string]string{"foo.t": "foo.t"}, false},
		{"foo=foo.t", map[string]string{"foo": "foo.t"}, false},
		{"dir/foo.t", map[string]string{"dir/foo.t": "dir/foo.t"}, false},
		{"foo=dir/foo.t", map[string]string{"foo": "dir/foo.t"}, false},
		{"dir/", map[string]string{"dir/foo.t": "dir/foo.t", "dir/bar.t": "dir/bar.t"}, false},
		{"t=dir/", map[string]string{"t/foo.t": "dir/foo.t", "t/bar.t": "dir/bar.t"}, false},
	}

	for _, d := range testdata {
		nested := templateAliases{}
		err := parseTemplateArg(d.arg, nested)
		if d.err {
			assert.Error(t, err, d.arg)
		} else {
			assert.NoError(t, err, d.arg)
			assert.Equal(t, templateAliases(d.expected), nested, d.arg)
		}
	}
}

func TestParseTemplateArgs(t *testing.T) {
	fs = afero.NewMemMapFs()
	afero.WriteFile(fs, "foo.t", []byte("hi"), 0600)
	_ = fs.MkdirAll("dir", 0755)
	afero.WriteFile(fs, "dir/foo.t", []byte("hi"), 0600)
	afero.WriteFile(fs, "dir/bar.t", []byte("hi"), 0600)

	args := []string{"foo.t",
		"foo=foo.t",
		"bar=dir/foo.t",
		"dir/",
		"t=dir/",
	}

	expected := map[string]string{
		"foo.t":     "foo.t",
		"foo":       "foo.t",
		"bar":       "dir/foo.t",
		"dir/foo.t": "dir/foo.t",
		"dir/bar.t": "dir/bar.t",
		"t/foo.t":   "dir/foo.t",
		"t/bar.t":   "dir/bar.t",
	}

	nested, err := parseTemplateArgs(args)
	assert.NoError(t, err)
	assert.Equal(t, templateAliases(expected), nested)

	_, err = parseTemplateArgs([]string{"bogus.t"})
	assert.Error(t, err)
}

func TestSimpleNamer(t *testing.T) {
	n := simpleNamer("out/")
	out, err := n("file")
	assert.NoError(t, err)
	expected := filepath.FromSlash("out/file")
	assert.Equal(t, expected, out)
}

func TestMappingNamer(t *testing.T) {
	g := &gomplate{funcMap: map[string]interface{}{
		"foo": func() string { return "foo" },
	}}
	n := mappingNamer("out/{{ .in }}", g)
	out, err := n("file")
	assert.NoError(t, err)
	expected := filepath.FromSlash("out/file")
	assert.Equal(t, expected, out)

	n = mappingNamer("out/{{ foo }}{{ .in }}", g)
	out, err = n("file")
	assert.NoError(t, err)
	expected = filepath.FromSlash("out/foofile")
	assert.Equal(t, expected, out)
}
