package gomplate

import (
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
)

func TestTplFunc(t *testing.T) {
	g := &gomplate{
		leftDelim:  "{{",
		rightDelim: "}}",
		funcMap:    template.FuncMap{},
	}
	tmpl := &tplate{name: "root", contents: "foo"}
	tmpl.toGoTemplate(g)

	testdata := []string{
		"{{ print `hello world`}}",
		"{{ tpl \"{{ print `hello world`}}\"}}",
		"{{ tpl \"{{ tpl \\\"{{ print `hello world`}}\\\"}}\"}}",
	}
	for _, d := range testdata {
		out, err := g.tpl(d)
		assert.NoError(t, err, d)
		assert.Equal(t, "hello world", out)
	}
}

func TestParseArgs(t *testing.T) {
	name, in, ctx, err := parseArgs("foo")
	assert.NoError(t, err)
	assert.Equal(t, "<inline>", name)
	assert.Equal(t, "foo", in)
	assert.EqualValues(t, &context{}, ctx)

	_, _, _, err = parseArgs(42)
	assert.Error(t, err)

	_, _, _, err = parseArgs()
	assert.Error(t, err)

	_, _, _, err = parseArgs("", "", 42, "")
	assert.Error(t, err)

	_, _, _, err = parseArgs("", 42, 42)
	assert.Error(t, err)

	name, in, ctx, err = parseArgs("foo", "bar")
	assert.NoError(t, err)
	assert.Equal(t, "foo", name)
	assert.Equal(t, "bar", in)
	assert.EqualValues(t, &context{}, ctx)

	c := map[string]string{"one": "two"}
	name, in, ctx, err = parseArgs("foo", c)
	assert.NoError(t, err)
	assert.Equal(t, "<inline>", name)
	assert.Equal(t, "foo", in)
	assert.Equal(t, c, ctx)

	name, in, ctx, err = parseArgs("foo", "bar", c)
	assert.NoError(t, err)
	assert.Equal(t, "foo", name)
	assert.Equal(t, "bar", in)
	assert.Equal(t, c, ctx)
}
