package gomplate

import (
	"net/url"
	"os"
	"testing"

	"github.com/hairyhenderson/gomplate/data"

	"github.com/stretchr/testify/assert"
)

func TestEnvMapifiesEnvironment(t *testing.T) {
	c := &context{}
	env := c.Env()
	assert.Equal(t, env["USER"], os.Getenv("USER"))
}

func TestEnvGetsUpdatedEnvironment(t *testing.T) {
	c := &context{}
	assert.Empty(t, c.Env()["FOO"])
	assert.NoError(t, os.Setenv("FOO", "foo"))
	assert.Equal(t, c.Env()["FOO"], "foo")
}

func TestCreateContext(t *testing.T) {
	c, err := createContext(nil, nil)
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
	c, err = createContext([]string{"foo=" + fooURL}, d)
	assert.NoError(t, err)
	assert.IsType(t, &context{}, c)
	ctx := c.(*context)
	ds := ((*ctx)["foo"]).(map[string]interface{})
	assert.Equal(t, "bar", ds["foo"])

	os.Setenv("bar", "bar: baz")
	defer os.Unsetenv("bar")
	c, err = createContext([]string{".=" + barURL}, d)
	assert.NoError(t, err)
	assert.IsType(t, map[string]interface{}{}, c)
	ds = c.(map[string]interface{})
	assert.Equal(t, "baz", ds["bar"])
}

func TestParseAlias(t *testing.T) {
	testdata := map[string]string{
		"":        "",
		"foo":     "foo",
		"foo.bar": "foo",
		"a=b":     "a",
		".=foo":   ".",
	}
	for k, v := range testdata {
		assert.Equal(t, v, parseAlias(k))
	}
}
