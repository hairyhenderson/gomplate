package funcs

import (
	"context"
	"reflect"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateCollFuncs(t *testing.T) {
	t.Parallel()

	for i := range 10 {
		// Run this a bunch to catch race conditions
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			fmap := CreateCollFuncs(ctx)
			actual := fmap["coll"].(func() any)

			assert.Equal(t, ctx, actual().(*CollFuncs).ctx)
		})
	}
}

func TestFlatten(t *testing.T) {
	t.Parallel()

	c := CollFuncs{}

	_, err := c.Flatten()
	require.Error(t, err)

	_, err = c.Flatten(42)
	require.Error(t, err)

	out, err := c.Flatten([]any{1, []any{[]int{2}, 3}})
	require.NoError(t, err)
	assert.Equal(t, []any{1, 2, 3}, out)

	out, err = c.Flatten(1, []any{1, []any{[]int{2}, 3}})
	require.NoError(t, err)
	assert.Equal(t, []any{1, []int{2}, 3}, out)
}

func TestPick(t *testing.T) {
	t.Parallel()

	c := &CollFuncs{}

	_, err := c.Pick()
	require.Error(t, err)

	_, err = c.Pick("")
	require.Error(t, err)

	_, err = c.Pick("foo", nil)
	require.Error(t, err)

	_, err = c.Pick("foo", "bar")
	require.Error(t, err)

	_, err = c.Pick(map[string]any{}, "foo", "bar", map[string]any{})
	require.Error(t, err)

	in := map[string]any{
		"foo": "bar",
		"bar": true,
		"":    "baz",
	}
	out, err := c.Pick("baz", in)
	require.NoError(t, err)
	assert.Equal(t, map[string]any{}, out)

	expected := map[string]any{
		"foo": "bar",
		"bar": true,
	}
	out, err = c.Pick("foo", "bar", in)
	require.NoError(t, err)
	assert.Equal(t, expected, out)

	expected = map[string]any{
		"": "baz",
	}
	out, err = c.Pick("", in)
	require.NoError(t, err)
	assert.Equal(t, expected, out)

	out, err = c.Pick("foo", "bar", "", in)
	require.NoError(t, err)
	assert.Equal(t, in, out)

	t.Run("supports slice key", func(t *testing.T) {
		t.Parallel()

		in := map[string]any{
			"foo": "bar",
			"bar": true,
			"":    "baz",
		}
		out, err := c.Pick([]string{"foo", "bar"}, in)
		require.NoError(t, err)
		assert.Equal(t, map[string]any{"foo": "bar", "bar": true}, out)
	})
}

func TestOmit(t *testing.T) {
	t.Parallel()

	c := &CollFuncs{}

	_, err := c.Omit()
	require.Error(t, err)

	_, err = c.Omit("")
	require.Error(t, err)

	_, err = c.Omit("foo", nil)
	require.Error(t, err)

	_, err = c.Omit("foo", "bar")
	require.Error(t, err)

	_, err = c.Omit(map[string]any{}, "foo", "bar", map[string]any{})
	require.Error(t, err)

	in := map[string]any{
		"foo": "bar",
		"bar": true,
		"":    "baz",
	}
	out, err := c.Omit("baz", in)
	require.NoError(t, err)
	assert.Equal(t, in, out)

	expected := map[string]any{
		"foo": "bar",
		"bar": true,
	}
	out, err = c.Omit("", in)
	require.NoError(t, err)
	assert.Equal(t, expected, out)

	expected = map[string]any{
		"": "baz",
	}
	out, err = c.Omit("foo", "bar", in)
	require.NoError(t, err)
	assert.Equal(t, expected, out)

	out, err = c.Omit("foo", "bar", "", in)
	require.NoError(t, err)
	assert.Equal(t, map[string]any{}, out)

	t.Run("supports slice of strings", func(t *testing.T) {
		t.Parallel()

		in := map[string]any{
			"foo": "bar",
			"bar": true,
			"":    "baz",
		}
		out, err := c.Omit([]string{"foo", "bar"}, in)
		require.NoError(t, err)
		assert.Equal(t, map[string]any{"": "baz"}, out)
	})

	t.Run("supports slice of any", func(t *testing.T) {
		t.Parallel()

		in := map[string]any{
			"foo": "bar",
			"bar": true,
			"":    "baz",
		}
		out, err := c.Omit([]any{"foo", "bar"}, in)
		require.NoError(t, err)
		assert.Equal(t, map[string]any{"": "baz"}, out)
	})
}

func TestGoSlice(t *testing.T) {
	t.Parallel()

	c := &CollFuncs{}

	in := reflect.ValueOf(nil)
	_, err := c.GoSlice(in)
	require.Error(t, err)

	in = reflect.ValueOf(42)
	_, err = c.GoSlice(in)
	require.Error(t, err)

	// invalid index type
	in = reflect.ValueOf([]any{1})
	_, err = c.GoSlice(in, reflect.ValueOf([]any{[]int{2}}))
	require.Error(t, err)

	// valid slice, no slicing
	in = reflect.ValueOf([]int{1})
	out, err := c.GoSlice(in)
	require.NoError(t, err)
	assert.Equal(t, reflect.TypeOf([]int{}), out.Type())
	assert.Equal(t, []int{1}, out.Interface())

	// valid slice, slicing
	in = reflect.ValueOf([]string{"foo", "bar", "baz"})
	out, err = c.GoSlice(in, reflect.ValueOf(1), reflect.ValueOf(3))
	require.NoError(t, err)
	assert.Equal(t, reflect.TypeOf([]string{}), out.Type())
	assert.Equal(t, []string{"bar", "baz"}, out.Interface())
}

func TestCollFuncs_Set(t *testing.T) {
	t.Parallel()

	c := &CollFuncs{}

	m := map[string]any{"foo": "bar"}
	out, err := c.Set("foo", "baz", m)
	require.NoError(t, err)
	assert.Equal(t, map[string]any{"foo": "baz"}, out)

	// m was modified so foo is now baz
	out, err = c.Set("bar", "baz", m)
	require.NoError(t, err)
	assert.Equal(t, map[string]any{"foo": "baz", "bar": "baz"}, out)
}

func TestCollFuncs_Unset(t *testing.T) {
	t.Parallel()

	c := &CollFuncs{}

	m := map[string]any{"foo": "bar"}
	out, err := c.Unset("foo", m)
	require.NoError(t, err)
	assert.Empty(t, out)

	// no-op
	out, err = c.Unset("bar", m)
	require.NoError(t, err)
	assert.Empty(t, out)
}
