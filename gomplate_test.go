package gomplate

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"text/template"

	"github.com/hairyhenderson/gomplate/v3/aws"
	"github.com/hairyhenderson/gomplate/v3/conv"
	"github.com/hairyhenderson/gomplate/v3/data"
	"github.com/hairyhenderson/gomplate/v3/env"

	"github.com/stretchr/testify/assert"
)

func testTemplate(t *testing.T, g *gomplate, tmpl string) string {
	t.Helper()

	var out bytes.Buffer
	err := g.runTemplate(context.Background(), &tplate{name: "testtemplate", contents: tmpl, target: &out})
	assert.NoError(t, err)

	return out.String()
}

func TestGetenvTemplates(t *testing.T) {
	g := &gomplate{
		funcMap: template.FuncMap{
			"getenv": env.Getenv,
			"bool":   conv.Bool,
		},
	}
	assert.Empty(t, testTemplate(t, g, `{{getenv "BLAHBLAHBLAH"}}`))
	assert.Equal(t, os.Getenv("USER"), testTemplate(t, g, `{{getenv "USER"}}`))
	assert.Equal(t, "default value", testTemplate(t, g, `{{getenv "BLAHBLAHBLAH" "default value"}}`))
}

func TestBoolTemplates(t *testing.T) {
	g := &gomplate{
		funcMap: template.FuncMap{
			"bool": conv.Bool,
		},
	}
	assert.Equal(t, "true", testTemplate(t, g, `{{bool "true"}}`))
	assert.Equal(t, "false", testTemplate(t, g, `{{bool "false"}}`))
	assert.Equal(t, "false", testTemplate(t, g, `{{bool "foo"}}`))
	assert.Equal(t, "false", testTemplate(t, g, `{{bool ""}}`))
}

func TestEc2MetaTemplates(t *testing.T) {
	createGomplate := func(data map[string]string, region string) *gomplate {
		ec2meta := aws.MockEC2Meta(data, nil, region)
		return &gomplate{funcMap: template.FuncMap{"ec2meta": ec2meta.Meta}}
	}

	g := createGomplate(nil, "")
	assert.Equal(t, "", testTemplate(t, g, `{{ec2meta "foo"}}`))
	assert.Equal(t, "default", testTemplate(t, g, `{{ec2meta "foo" "default"}}`))

	g = createGomplate(map[string]string{"instance-id": "i-1234"}, "")
	assert.Equal(t, "i-1234", testTemplate(t, g, `{{ec2meta "instance-id"}}`))
	assert.Equal(t, "i-1234", testTemplate(t, g, `{{ec2meta "instance-id" "default"}}`))
}

func TestEc2MetaTemplates_WithJSON(t *testing.T) {
	ec2meta := aws.MockEC2Meta(map[string]string{"obj": `"foo": "bar"`}, map[string]string{"obj": `"foo": "baz"`}, "")

	g := &gomplate{
		funcMap: template.FuncMap{
			"ec2meta":    ec2meta.Meta,
			"ec2dynamic": ec2meta.Dynamic,
			"json":       data.JSON,
		},
	}

	assert.Equal(t, "bar", testTemplate(t, g, `{{ (ec2meta "obj" | json).foo }}`))
	assert.Equal(t, "baz", testTemplate(t, g, `{{ (ec2dynamic "obj" | json).foo }}`))
}

func TestJSONArrayTemplates(t *testing.T) {
	g := &gomplate{
		funcMap: template.FuncMap{
			"jsonArray": data.JSONArray,
		},
	}

	assert.Equal(t, "[foo bar]", testTemplate(t, g, `{{jsonArray "[\"foo\",\"bar\"]"}}`))
	assert.Equal(t, "bar", testTemplate(t, g, `{{ index (jsonArray "[\"foo\",\"bar\"]") 1 }}`))
}

func TestYAMLTemplates(t *testing.T) {
	g := &gomplate{
		funcMap: template.FuncMap{
			"yaml":      data.YAML,
			"yamlArray": data.YAMLArray,
		},
	}

	assert.Equal(t, "bar", testTemplate(t, g, `{{(yaml "foo: bar").foo}}`))
	assert.Equal(t, "[foo bar]", testTemplate(t, g, `{{yamlArray "- foo\n- bar\n"}}`))
	assert.Equal(t, "bar", testTemplate(t, g, `{{ index (yamlArray "[\"foo\",\"bar\"]") 1 }}`))
}

func TestSliceTemplates(t *testing.T) {
	g := &gomplate{
		funcMap: template.FuncMap{
			"slice": conv.Slice,
		},
	}
	assert.Equal(t, "foo", testTemplate(t, g, `{{index (slice "foo") 0}}`))
	assert.Equal(t, `[foo bar 42]`, testTemplate(t, g, `{{slice "foo" "bar" 42}}`))
	assert.Equal(t, `helloworld`, testTemplate(t, g, `{{range slice "hello" "world"}}{{.}}{{end}}`))
}

func TestHasTemplate(t *testing.T) {
	g := &gomplate{
		funcMap: template.FuncMap{
			"yaml": data.YAML,
			"has":  conv.Has,
		},
	}
	assert.Equal(t, "true", testTemplate(t, g, `{{has ("foo:\n  bar: true" | yaml) "foo"}}`))
	assert.Equal(t, "true", testTemplate(t, g, `{{has ("foo:\n  bar: true" | yaml).foo "bar"}}`))
	assert.Equal(t, "false", testTemplate(t, g, `{{has ("foo: true" | yaml) "bah"}}`))
	tmpl := `{{- $data := yaml "foo: bar\nbaz: qux\n" }}
{{- if (has $data "baz") }}
{{- $data.baz }}
{{- end }}`
	assert.Equal(t, "qux", testTemplate(t, g, tmpl))
	tmpl = `{{- $data := yaml "foo: bar\nbaz: qux\n" }}
{{- if (has $data "quux") }}
{{- $data.quux }}
{{- else }}
{{- $data.foo }}
{{- end }}`
	assert.Equal(t, "bar", testTemplate(t, g, tmpl))
}

func TestCustomDelim(t *testing.T) {
	g := &gomplate{
		leftDelim:  "[",
		rightDelim: "]",
		funcMap:    template.FuncMap{},
	}
	assert.Equal(t, "hi", testTemplate(t, g, `[print "hi"]`))
}

func TestRunTemplates(t *testing.T) {
	buf := &bytes.Buffer{}
	config := &Config{Input: "foo", OutputFiles: []string{"-"}, Out: buf}
	err := RunTemplates(config)
	assert.NoError(t, err)
	assert.Equal(t, "foo", buf.String())
	assert.Equal(t, 1, Metrics.TemplatesGathered)
	assert.Equal(t, 1, Metrics.TemplatesProcessed)
	assert.Equal(t, 0, Metrics.Errors)
}

func TestSimpleNamer(t *testing.T) {
	n := simpleNamer("out/")
	out, err := n(context.Background(), "file")
	assert.NoError(t, err)
	expected := filepath.FromSlash("out/file")
	assert.Equal(t, expected, out)
}

func TestMappingNamer(t *testing.T) {
	ctx := context.Background()
	g := &gomplate{funcMap: map[string]interface{}{
		"foo": func() string { return "foo" },
	}}
	n := mappingNamer("out/{{ .in }}", g)
	out, err := n(ctx, "file")
	assert.NoError(t, err)
	expected := filepath.FromSlash("out/file")
	assert.Equal(t, expected, out)

	n = mappingNamer("out/{{ foo }}{{ .in }}", g)
	out, err = n(ctx, "file")
	assert.NoError(t, err)
	expected = filepath.FromSlash("out/foofile")
	assert.Equal(t, expected, out)
}
