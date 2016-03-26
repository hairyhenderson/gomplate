package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"text/template"

	"github.com/hairyhenderson/gomplate/aws"
	"github.com/stretchr/testify/assert"
)

func testTemplate(g *Gomplate, template string) string {
	in := strings.NewReader(template)
	var out bytes.Buffer
	g.RunTemplate(in, &out)
	return strings.TrimSpace(out.String())
}

func TestGetenvTemplates(t *testing.T) {
	g := NewGomplate()
	assert.Empty(t, testTemplate(g, `{{getenv "BLAHBLAHBLAH"}}`))
	assert.Equal(t, os.Getenv("USER"), testTemplate(g, `{{getenv "USER"}}`))
	assert.Equal(t, "default value", testTemplate(g, `{{getenv "BLAHBLAHBLAH" "default value"}}`))
}

func TestBoolTemplates(t *testing.T) {
	g := NewGomplate()
	assert.Equal(t, "true", testTemplate(g, `{{bool "true"}}`))
	assert.Equal(t, "false", testTemplate(g, `{{bool "false"}}`))
	assert.Equal(t, "false", testTemplate(g, `{{bool "foo"}}`))
	assert.Equal(t, "false", testTemplate(g, `{{bool ""}}`))
}

func TestEc2MetaTemplates_MissingKey(t *testing.T) {
	server, ec2meta := aws.MockServer(404, "")
	defer server.Close()
	g := &Gomplate{
		funcMap: template.FuncMap{
			"ec2meta": ec2meta.Ec2meta,
		},
	}

	assert.Equal(t, "", testTemplate(g, `{{ec2meta "foo"}}`))
	assert.Equal(t, "default", testTemplate(g, `{{ec2meta "foo" "default"}}`))
}

func TestEc2MetaTemplates_ValidKey(t *testing.T) {
	server, ec2meta := aws.MockServer(200, "i-1234")
	defer server.Close()
	g := &Gomplate{
		funcMap: template.FuncMap{
			"ec2meta": ec2meta.Ec2meta,
		},
	}

	assert.Equal(t, "i-1234", testTemplate(g, `{{ec2meta "instance-id"}}`))
	assert.Equal(t, "i-1234", testTemplate(g, `{{ec2meta "instance-id" "default"}}`))
}

func TestEc2MetaTemplates_WithJSON(t *testing.T) {
	server, ec2meta := aws.MockServer(200, `{"foo":"bar"}`)
	defer server.Close()
	ty := new(TypeConv)
	g := &Gomplate{
		funcMap: template.FuncMap{
			"ec2meta":    ec2meta.Ec2meta,
			"ec2dynamic": ec2meta.Ec2dynamic,
			"json":       ty.JSON,
		},
	}

	assert.Equal(t, "bar", testTemplate(g, `{{ (ec2meta "obj" | json).foo }}`))
	assert.Equal(t, "bar", testTemplate(g, `{{ (ec2dynamic "obj" | json).foo }}`))
}
