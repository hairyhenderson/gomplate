package tmpl

import (
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
)

func TestInline(t *testing.T) {
	testdata := []string{
		"{{ print `hello world`}}",
		"{{ tpl \"{{ print `hello world`}}\"}}",
		"{{ tpl \"{{ tpl \\\"{{ print `hello world`}}\\\"}}\"}}",
	}
	tmpl := &Template{
		defaultCtx: map[string]string{},
		root:       template.New("root"),
	}
	tmpl.root.Funcs(template.FuncMap{
		"tpl": tmpl.Inline,
	})
	for _, d := range testdata {
		out, err := tmpl.Inline(d)
		assert.NoError(t, err, d)
		assert.Equal(t, "hello world", out)
	}
}

func TestParseArgs(t *testing.T) {
	defaultCtx := map[string]string{"hello": "world"}
	tmpl := New(nil, defaultCtx)
	name, in, ctx, err := tmpl.parseArgs("foo")
	assert.NoError(t, err)
	assert.Equal(t, "<inline>", name)
	assert.Equal(t, "foo", in)
	assert.Equal(t, defaultCtx, ctx)

	_, _, _, err = tmpl.parseArgs(42)
	assert.Error(t, err)

	_, _, _, err = tmpl.parseArgs()
	assert.Error(t, err)

	_, _, _, err = tmpl.parseArgs("", "", 42, "")
	assert.Error(t, err)

	_, _, _, err = tmpl.parseArgs("", 42, 42)
	assert.Error(t, err)

	name, in, ctx, err = tmpl.parseArgs("foo", "bar")
	assert.NoError(t, err)
	assert.Equal(t, "foo", name)
	assert.Equal(t, "bar", in)
	assert.Equal(t, defaultCtx, ctx)

	c := map[string]string{"one": "two"}
	name, in, ctx, err = tmpl.parseArgs("foo", c)
	assert.NoError(t, err)
	assert.Equal(t, "<inline>", name)
	assert.Equal(t, "foo", in)
	assert.Equal(t, c, ctx)

	name, in, ctx, err = tmpl.parseArgs("foo", "bar", c)
	assert.NoError(t, err)
	assert.Equal(t, "foo", name)
	assert.Equal(t, "bar", in)
	assert.Equal(t, c, ctx)
}

func TestExec(t *testing.T) {
	root := template.New("root")
	t1 := root.New("T1")
	t1.Parse("hello, {{ . }}")
	tmpl := &Template{
		defaultCtx: map[string]string{"foo": "bar"},
		root:       root,
	}

	out, err := tmpl.Exec("T1")
	assert.NoError(t, err)
	assert.Equal(t, "hello, map[foo:bar]", out)

	out, err = tmpl.Exec("T1", "world")
	assert.NoError(t, err)
	assert.Equal(t, "hello, world", out)

	_, err = tmpl.Exec("bogus")
	assert.Error(t, err)
}
