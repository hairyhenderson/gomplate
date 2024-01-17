package gomplate

import (
	"context"
	"net/url"
	"os"
	"testing"

	"github.com/hairyhenderson/go-fsimpl"
	"github.com/hairyhenderson/gomplate/v4/data"
	"github.com/hairyhenderson/gomplate/v4/internal/datafs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvMapifiesEnvironment(t *testing.T) {
	c := &tmplctx{}
	env := c.Env()
	assert.Equal(t, env["USER"], os.Getenv("USER"))
}

func TestEnvGetsUpdatedEnvironment(t *testing.T) {
	c := &tmplctx{}
	assert.Empty(t, c.Env()["FOO"])
	t.Setenv("FOO", "foo")
	assert.Equal(t, c.Env()["FOO"], "foo")
}

func TestCreateContext(t *testing.T) {
	ctx := context.Background()
	c, err := createTmplContext(ctx, nil, nil)
	require.NoError(t, err)
	assert.Empty(t, c)

	fsmux := fsimpl.NewMux()
	fsmux.Add(datafs.EnvFS)

	ctx = datafs.ContextWithFSProvider(ctx, fsmux)

	fooURL := "env:///foo?type=application/yaml"
	barURL := "env:///bar?type=application/yaml"
	uf, _ := url.Parse(fooURL)
	ub, _ := url.Parse(barURL)
	//nolint:staticcheck
	d := &data.Data{
		Sources: map[string]*data.Source{
			"foo": {URL: uf},
			".":   {URL: ub},
		},
	}
	t.Setenv("foo", "foo: bar")
	c, err = createTmplContext(ctx, []string{"foo"}, d)
	require.NoError(t, err)
	assert.IsType(t, &tmplctx{}, c)
	tctx := c.(*tmplctx)
	ds := ((*tctx)["foo"]).(map[string]interface{})
	assert.Equal(t, "bar", ds["foo"])

	t.Setenv("bar", "bar: baz")
	c, err = createTmplContext(ctx, []string{"."}, d)
	require.NoError(t, err)
	assert.IsType(t, map[string]interface{}{}, c)
	ds = c.(map[string]interface{})
	assert.Equal(t, "baz", ds["bar"])
}
