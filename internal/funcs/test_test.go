package funcs

import (
	"context"
	"strconv"
	"testing"

	"github.com/hairyhenderson/gomplate/v5/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateTestFuncs(t *testing.T) {
	t.Parallel()

	for i := range 10 {
		// Run this a bunch to catch race conditions
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			fmap := CreateTestFuncs(ctx)
			actual := fmap["test"].(func() any)

			assert.Equal(t, ctx, actual().(*TestFuncs).ctx)
		})
	}
}

func TestAssert(t *testing.T) {
	t.Parallel()

	f := TestFuncs{ctx: context.Background()}
	_, err := f.Assert()
	require.Error(t, err)

	_, err = f.Assert(false)
	require.Error(t, err)

	_, err = f.Assert(true)
	require.NoError(t, err)

	_, err = f.Assert("foo", true)
	require.NoError(t, err)

	_, err = f.Assert("foo", "false")
	require.EqualError(t, err, "assertion failed: foo")
}

func TestRequired(t *testing.T) {
	t.Parallel()

	f := TestFuncs{ctx: context.Background()}
	errMsg := "can not render template: a required value was not set"
	v, err := f.Required("")
	require.Error(t, err)
	require.EqualError(t, err, errMsg)
	assert.Nil(t, v)

	v, err = f.Required(nil)
	require.Error(t, err)
	require.EqualError(t, err, errMsg)
	assert.Nil(t, v)

	errMsg = "hello world"
	v, err = f.Required(errMsg, nil)
	require.Error(t, err)
	require.EqualError(t, err, errMsg)
	assert.Nil(t, v)

	v, err = f.Required(42, nil)
	require.Error(t, err)
	require.EqualError(t, err, "at <1>: expected string; found int")
	assert.Nil(t, v)

	v, err = f.Required()
	require.Error(t, err)
	require.EqualError(t, err, "wrong number of args: want 1 or 2, got 0")
	assert.Nil(t, v)

	v, err = f.Required("", 2, 3)
	require.Error(t, err)
	require.EqualError(t, err, "wrong number of args: want 1 or 2, got 3")
	assert.Nil(t, v)

	v, err = f.Required(0)
	require.NoError(t, err)
	assert.Zero(t, v)

	v, err = f.Required("foo")
	require.NoError(t, err)
	assert.Equal(t, "foo", v)
}

func TestTernary(t *testing.T) {
	t.Parallel()

	f := TestFuncs{ctx: context.Background()}
	testdata := []struct {
		tval, fval, b any
		expected      any
	}{
		{"foo", 42, false, 42},
		{"foo", 42, "yes", "foo"},
		{false, true, true, false},
	}
	for _, d := range testdata {
		assert.Equal(t, d.expected, f.Ternary(d.tval, d.fval, d.b))
	}
}

func TestKind(t *testing.T) {
	t.Parallel()

	f := TestFuncs{ctx: context.Background()}
	testdata := []struct {
		arg      any
		expected string
	}{
		{"foo", "string"},
		{nil, "invalid"},
		{false, "bool"},
		{[]string{"foo", "bar"}, "slice"},
		{map[string]string{"foo": "bar"}, "map"},
		{42, "int"},
		{42.0, "float64"},
		{uint(42), "uint"},
		{struct{}{}, "struct"},
	}
	for _, d := range testdata {
		assert.Equal(t, d.expected, f.Kind(d.arg))
	}
}

func TestIsKind(t *testing.T) {
	t.Parallel()

	f := TestFuncs{ctx: context.Background()}
	truedata := []struct {
		arg  any
		kind string
	}{
		{"foo", "string"},
		{nil, "invalid"},
		{false, "bool"},
		{[]string{"foo", "bar"}, "slice"},
		{map[string]string{"foo": "bar"}, "map"},
		{42, "int"},
		{42.0, "float64"},
		{uint(42), "uint"},
		{struct{}{}, "struct"},
		{42.0, "number"},
		{42, "number"},
		{uint32(64000), "number"},
		{complex128(64000), "number"},
	}
	for _, d := range truedata {
		assert.True(t, f.IsKind(d.kind, d.arg))
	}

	falsedata := []struct {
		arg  any
		kind string
	}{
		{"foo", "bool"},
		{nil, "struct"},
		{false, "string"},
		{[]string{"foo", "bar"}, "map"},
		{map[string]string{"foo": "bar"}, "int"},
		{42, "int64"},
		{42.0, "float32"},
		{uint(42), "int"},
		{struct{}{}, "interface"},
	}
	for _, d := range falsedata {
		assert.False(t, f.IsKind(d.kind, d.arg))
	}
}

func TestJSONSchema(t *testing.T) {
	t.Parallel()

	schema := map[string]any{
		"type":     "object",
		"required": []any{"name"},
		"properties": map[string]any{
			"name": map[string]any{"type": "string"},
		},
	}
	valid := map[string]any{"name": "foo"}

	f := TestFuncs{ctx: context.Background()}
	_, err := f.JSONSchema(schema, valid)
	require.ErrorContains(t, err, "experimental")

	f = TestFuncs{ctx: config.SetExperimental(context.Background())}
	v, err := f.JSONSchema(schema, valid)
	require.NoError(t, err)
	assert.Equal(t, valid, v)

	v, err = f.JSONSchema(`{"type": "string"}`, "foo")
	require.NoError(t, err)
	assert.Equal(t, "foo", v)

	v, err = f.JSONSchema(schema, map[string]any{"name": 42})
	require.ErrorContains(t, err, "schema validation failed")
	assert.Nil(t, v)
}
