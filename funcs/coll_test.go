package funcs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlatten(t *testing.T) {
	c := CollNS()

	_, err := c.Flatten()
	assert.Error(t, err)

	_, err = c.Flatten(42)
	assert.Error(t, err)

	out, err := c.Flatten([]interface{}{1, []interface{}{[]int{2}, 3}})
	assert.NoError(t, err)
	assert.EqualValues(t, []interface{}{1, 2, 3}, out)

	out, err = c.Flatten(1, []interface{}{1, []interface{}{[]int{2}, 3}})
	assert.NoError(t, err)
	assert.EqualValues(t, []interface{}{1, []int{2}, 3}, out)
}

func TestPick(t *testing.T) {
	c := &CollFuncs{}

	_, err := c.Pick()
	assert.Error(t, err)

	_, err = c.Pick("")
	assert.Error(t, err)

	_, err = c.Pick("foo", nil)
	assert.Error(t, err)

	_, err = c.Pick("foo", "bar")
	assert.Error(t, err)

	_, err = c.Pick(map[string]interface{}{}, "foo", "bar", map[string]interface{}{})
	assert.Error(t, err)

	in := map[string]interface{}{
		"foo": "bar",
		"bar": true,
		"":    "baz",
	}
	out, err := c.Pick("baz", in)
	assert.NoError(t, err)
	assert.EqualValues(t, map[string]interface{}{}, out)

	expected := map[string]interface{}{
		"foo": "bar",
		"bar": true,
	}
	out, err = c.Pick("foo", "bar", in)
	assert.NoError(t, err)
	assert.EqualValues(t, expected, out)

	expected = map[string]interface{}{
		"": "baz",
	}
	out, err = c.Pick("", in)
	assert.NoError(t, err)
	assert.EqualValues(t, expected, out)

	out, err = c.Pick("foo", "bar", "", in)
	assert.NoError(t, err)
	assert.EqualValues(t, in, out)
}

func TestOmit(t *testing.T) {
	c := &CollFuncs{}

	_, err := c.Omit()
	assert.Error(t, err)

	_, err = c.Omit("")
	assert.Error(t, err)

	_, err = c.Omit("foo", nil)
	assert.Error(t, err)

	_, err = c.Omit("foo", "bar")
	assert.Error(t, err)

	_, err = c.Omit(map[string]interface{}{}, "foo", "bar", map[string]interface{}{})
	assert.Error(t, err)

	in := map[string]interface{}{
		"foo": "bar",
		"bar": true,
		"":    "baz",
	}
	out, err := c.Omit("baz", in)
	assert.NoError(t, err)
	assert.EqualValues(t, in, out)

	expected := map[string]interface{}{
		"foo": "bar",
		"bar": true,
	}
	out, err = c.Omit("", in)
	assert.NoError(t, err)
	assert.EqualValues(t, expected, out)

	expected = map[string]interface{}{
		"": "baz",
	}
	out, err = c.Omit("foo", "bar", in)
	assert.NoError(t, err)
	assert.EqualValues(t, expected, out)

	out, err = c.Omit("foo", "bar", "", in)
	assert.NoError(t, err)
	assert.EqualValues(t, map[string]interface{}{}, out)
}
