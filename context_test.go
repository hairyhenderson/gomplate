package gomplate

import (
	"context"
	"net/url"
	"os"
	"testing"

	"github.com/hairyhenderson/gomplate/v3/data"
	"github.com/hairyhenderson/gomplate/v3/internal/config"

	"github.com/stretchr/testify/assert"
)

func TestEnvMapifiesEnvironment(t *testing.T) {
	c := &tmplctx{}
	env := c.Env()
	assert.Equal(t, env["USER"], os.Getenv("USER"))
}

func TestEnvGetsUpdatedEnvironment(t *testing.T) {
	c := &tmplctx{}
	assert.Empty(t, c.Env()["FOO"])
	assert.NoError(t, os.Setenv("FOO", "foo"))
	assert.Equal(t, c.Env()["FOO"], "foo")
}

func TestCreateContext(t *testing.T) {
	ctx := context.Background()
	c, err := createTmplContext(ctx, nil, nil)
	assert.NoError(t, err)
	assert.Empty(t, c)

	fooURL := "env:///foo?type=application/yaml"
	barURL := "env:///bar?type=application/yaml"
	uf, _ := url.Parse(fooURL)
	ub, _ := url.Parse(barURL)
	d := &data.Data{
		Sources: map[string]*data.Source{
			"foo": {URL: uf},
			".":   {URL: ub},
		},
	}
	os.Setenv("foo", "foo: bar")
	defer os.Unsetenv("foo")
	c, err = createTmplContext(ctx, map[string]config.DataSource{"foo": {URL: uf}}, d)
	assert.NoError(t, err)
	assert.IsType(t, &tmplctx{}, c)
	tctx := c.(*tmplctx)
	ds := ((*tctx)["foo"]).(map[string]interface{})
	assert.Equal(t, "bar", ds["foo"])

	os.Setenv("bar", "bar: baz")
	defer os.Unsetenv("bar")
	c, err = createTmplContext(ctx, map[string]config.DataSource{".": {URL: ub}}, d)
	assert.NoError(t, err)
	assert.IsType(t, map[string]interface{}{}, c)
	ds = c.(map[string]interface{})
	assert.Equal(t, "baz", ds["bar"])
}
