package gomplate

import (
	"context"
	"net/url"
	"os"
	"testing"

	"github.com/hairyhenderson/go-fsimpl"
	"github.com/hairyhenderson/gomplate/v5/internal/datafs"

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
	assert.Equal(t, "foo", c.Env()["FOO"])
}

func TestCreateContext(t *testing.T) {
	ctx := context.Background()
	reg := datafs.NewRegistry()
	sr := datafs.NewSourceReader(reg)

	c, err := createTmplContext(ctx, nil, sr)
	require.NoError(t, err)
	assert.Empty(t, c)

	fsmux := fsimpl.NewMux()
	fsmux.Add(datafs.EnvFS)

	ctx = datafs.ContextWithFSProvider(ctx, fsmux)

	fooURL := "env:///foo?type=application/yaml"
	barURL := "env:///bar?type=application/yaml"
	uf, _ := url.Parse(fooURL)
	ub, _ := url.Parse(barURL)

	reg.Register("foo", DataSource{URL: uf})
	reg.Register(".", DataSource{URL: ub})

	t.Setenv("foo", "foo: bar")
	c, err = createTmplContext(ctx, []string{"foo"}, sr)
	require.NoError(t, err)
	assert.IsType(t, &tmplctx{}, c)
	tctx := c.(*tmplctx)
	ds := ((*tctx)["foo"]).(map[string]any)
	assert.Equal(t, "bar", ds["foo"])

	t.Setenv("bar", "bar: baz")
	c, err = createTmplContext(ctx, []string{"."}, sr)
	require.NoError(t, err)
	assert.IsType(t, map[string]any{}, c)
	ds = c.(map[string]any)
	assert.Equal(t, "baz", ds["bar"])
}
