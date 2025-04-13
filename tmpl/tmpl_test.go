package tmpl

import (
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		require.NoError(t, err, d)
		assert.Equal(t, "hello world", out)
	}
}

func TestParseArgs(t *testing.T) {
	defaultCtx := map[string]string{"hello": "world"}
	tmpl := New(nil, defaultCtx, "")
	name, in, ctx, err := tmpl.parseArgs("foo")
	require.NoError(t, err)
	assert.Equal(t, "<inline>", name)
	assert.Equal(t, "foo", in)
	assert.Equal(t, defaultCtx, ctx)

	_, _, _, err = tmpl.parseArgs(42)
	require.Error(t, err)

	_, _, _, err = tmpl.parseArgs()
	require.Error(t, err)

	_, _, _, err = tmpl.parseArgs("", "", 42, "")
	require.Error(t, err)

	_, _, _, err = tmpl.parseArgs("", 42, 42)
	require.Error(t, err)

	name, in, ctx, err = tmpl.parseArgs("foo", "bar")
	require.NoError(t, err)
	assert.Equal(t, "foo", name)
	assert.Equal(t, "bar", in)
	assert.Equal(t, defaultCtx, ctx)

	c := map[string]string{"one": "two"}
	name, in, ctx, err = tmpl.parseArgs("foo", c)
	require.NoError(t, err)
	assert.Equal(t, "<inline>", name)
	assert.Equal(t, "foo", in)
	assert.Equal(t, c, ctx)

	name, in, ctx, err = tmpl.parseArgs("foo", "bar", c)
	require.NoError(t, err)
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
	require.NoError(t, err)
	assert.Equal(t, "hello, map[foo:bar]", out)

	out, err = tmpl.Exec("T1", "world")
	require.NoError(t, err)
	assert.Equal(t, "hello, world", out)

	_, err = tmpl.Exec("bogus")
	require.Error(t, err)
}

func TestPath(t *testing.T) {
	tmpl := New(nil, nil, "")

	p, err := tmpl.Path()
	require.NoError(t, err)
	assert.Empty(t, p)

	tmpl = New(nil, nil, "foo")
	p, err = tmpl.Path()
	require.NoError(t, err)
	assert.Equal(t, "foo", p)
}

func TestPathDir(t *testing.T) {
	tmpl := New(nil, nil, "")

	p, err := tmpl.PathDir()
	require.NoError(t, err)
	assert.Empty(t, p)

	tmpl = New(nil, nil, "foo")
	p, err = tmpl.PathDir()
	require.NoError(t, err)
	assert.Equal(t, ".", p)

	tmpl = New(nil, nil, "foo/bar")
	p, err = tmpl.PathDir()
	require.NoError(t, err)
	assert.Equal(t, "foo", p)
}
