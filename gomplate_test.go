package gomplate

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"text/template"

	"github.com/hairyhenderson/gomplate/v5/conv"
	"github.com/hairyhenderson/gomplate/v5/env"
	"github.com/hairyhenderson/gomplate/v5/internal/aws"
	"github.com/hairyhenderson/gomplate/v5/internal/datafs"
	"github.com/hairyhenderson/gomplate/v5/internal/parsers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testTemplate(t *testing.T, tr *renderer, tmpl string) string {
	t.Helper()

	var out bytes.Buffer
	err := tr.Render(context.Background(), "testtemplate", tmpl, &out)
	require.NoError(t, err)

	return out.String()
}

func TestGetenvTemplates(t *testing.T) {
	tr := newRenderer(RenderOptions{
		Funcs: template.FuncMap{
			"getenv": env.Getenv,
			"bool":   conv.ToBool,
		},
	})
	assert.Empty(t, testTemplate(t, tr, `{{getenv "BLAHBLAHBLAH"}}`))
	assert.Equal(t, os.Getenv("USER"), testTemplate(t, tr, `{{getenv "USER"}}`))
	assert.Equal(t, "default value", testTemplate(t, tr, `{{getenv "BLAHBLAHBLAH" "default value"}}`))
}

func TestBoolTemplates(t *testing.T) {
	g := newRenderer(RenderOptions{
		Funcs: template.FuncMap{
			"bool": conv.ToBool,
		},
	})
	assert.Equal(t, "true", testTemplate(t, g, `{{bool "true"}}`))
	assert.Equal(t, "false", testTemplate(t, g, `{{bool "false"}}`))
	assert.Equal(t, "false", testTemplate(t, g, `{{bool "foo"}}`))
	assert.Equal(t, "false", testTemplate(t, g, `{{bool ""}}`))
}

func TestEc2MetaTemplates(t *testing.T) {
	createGomplate := func(data map[string]string, region string) *renderer {
		ec2meta := aws.MockEC2Meta(data, nil, region)
		return newRenderer(RenderOptions{
			Funcs: template.FuncMap{"ec2meta": func(key string, def ...string) (string, error) {
				return ec2meta.Meta(t.Context(), key, def...)
			}},
		})
	}

	g := createGomplate(nil, "")
	assert.Empty(t, testTemplate(t, g, `{{ec2meta "foo"}}`))
	assert.Equal(t, "default", testTemplate(t, g, `{{ec2meta "foo" "default"}}`))

	g = createGomplate(map[string]string{"instance-id": "i-1234"}, "")
	assert.Equal(t, "i-1234", testTemplate(t, g, `{{ec2meta "instance-id"}}`))
	assert.Equal(t, "i-1234", testTemplate(t, g, `{{ec2meta "instance-id" "default"}}`))
}

func TestEc2MetaTemplates_WithJSON(t *testing.T) {
	ec2meta := aws.MockEC2Meta(map[string]string{"obj": `"foo": "bar"`}, map[string]string{"obj": `"foo": "baz"`}, "")

	g := newRenderer(RenderOptions{
		Funcs: template.FuncMap{
			"ec2meta": func(key string, def ...string) (string, error) {
				return ec2meta.Meta(t.Context(), key, def...)
			},
			"ec2dynamic": func(key string, def ...string) (string, error) {
				return ec2meta.Dynamic(t.Context(), key, def...)
			},
			"json": parsers.JSON,
		},
	})

	assert.Equal(t, "bar", testTemplate(t, g, `{{ (ec2meta "obj" | json).foo }}`))
	assert.Equal(t, "baz", testTemplate(t, g, `{{ (ec2dynamic "obj" | json).foo }}`))
}

func TestJSONArrayTemplates(t *testing.T) {
	g := newRenderer(RenderOptions{
		Funcs: template.FuncMap{
			"jsonArray": parsers.JSONArray,
		},
	})

	assert.Equal(t, "[foo bar]", testTemplate(t, g, `{{jsonArray "[\"foo\",\"bar\"]"}}`))
	assert.Equal(t, "bar", testTemplate(t, g, `{{ index (jsonArray "[\"foo\",\"bar\"]") 1 }}`))
}

func TestYAMLTemplates(t *testing.T) {
	g := newRenderer(RenderOptions{
		Funcs: template.FuncMap{
			"yaml":      parsers.YAML,
			"yamlArray": parsers.YAMLArray,
		},
	})

	assert.Equal(t, "bar", testTemplate(t, g, `{{(yaml "foo: bar").foo}}`))
	assert.Equal(t, "[foo bar]", testTemplate(t, g, `{{yamlArray "- foo\n- bar\n"}}`))
	assert.Equal(t, "bar", testTemplate(t, g, `{{ index (yamlArray "[\"foo\",\"bar\"]") 1 }}`))
}

func TestHasTemplate(t *testing.T) {
	g := newRenderer(RenderOptions{
		Funcs: template.FuncMap{
			"yaml": parsers.YAML,
			"has":  conv.Has,
		},
	})
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

func TestMissingKey(t *testing.T) {
	tests := map[string]struct {
		MissingKey  string
		ExpectedOut string
	}{
		"missing-key = zero":    {MissingKey: "zero", ExpectedOut: "<no value>"},
		"missing-key = invalid": {MissingKey: "invalid", ExpectedOut: "<no value>"},
		"missing-key = default": {MissingKey: "default", ExpectedOut: "<no value>"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			g := newRenderer(RenderOptions{
				MissingKey: tt.MissingKey,
			})
			tmpl := `{{ .name }}`
			assert.Equal(t, tt.ExpectedOut, testTemplate(t, g, tmpl))
		})
	}
}

func TestCustomDelim(t *testing.T) {
	g := newRenderer(RenderOptions{
		LDelim: "[",
		RDelim: "]",
	})
	assert.Equal(t, "hi", testTemplate(t, g, `[print "hi"]`))
}

func TestSimpleNamer(t *testing.T) {
	n := simpleNamer("out/")
	out, err := n.Name(context.Background(), "file")
	require.NoError(t, err)
	expected := filepath.FromSlash("out/file")
	assert.Equal(t, expected, out)
}

func TestMappingNamer(t *testing.T) {
	ctx := context.Background()
	reg := datafs.NewRegistry()
	tr := &renderer{
		sr: datafs.NewSourceReader(reg),
		funcs: map[string]any{
			"foo": func() string { return "foo" },
		},
	}
	n := mappingNamer("out/{{ .in }}", tr)
	out, err := n.Name(ctx, "file")
	require.NoError(t, err)
	expected := filepath.FromSlash("out/file")
	assert.Equal(t, expected, out)

	n = mappingNamer("out/{{ foo }}{{ .in }}", tr)
	out, err = n.Name(ctx, "file")
	require.NoError(t, err)
	expected = filepath.FromSlash("out/foofile")
	assert.Equal(t, expected, out)
}
