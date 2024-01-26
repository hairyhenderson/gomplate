package funcs

import (
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateTestFuncs(t *testing.T) {
	t.Parallel()

	for i := 0; i < 10; i++ {
		// Run this a bunch to catch race conditions
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			fmap := CreateTestFuncs(ctx)
			actual := fmap["test"].(func() interface{})

			assert.Equal(t, ctx, actual().(*TestFuncs).ctx)
		})
	}
}

func TestAssert(t *testing.T) {
	t.Parallel()

	f := TestNS()
	_, err := f.Assert(false)
	assert.Error(t, err)

	_, err = f.Assert(true)
	require.NoError(t, err)

	_, err = f.Assert("foo", true)
	require.NoError(t, err)

	_, err = f.Assert("foo", "false")
	assert.EqualError(t, err, "assertion failed: foo")
}

func TestRequired(t *testing.T) {
	t.Parallel()

	f := TestNS()
	errMsg := "can not render template: a required value was not set"
	v, err := f.Required("")
	assert.Error(t, err)
	assert.EqualError(t, err, errMsg)
	assert.Nil(t, v)

	v, err = f.Required(nil)
	assert.Error(t, err)
	assert.EqualError(t, err, errMsg)
	assert.Nil(t, v)

	errMsg = "hello world"
	v, err = f.Required(errMsg, nil)
	assert.Error(t, err)
	assert.EqualError(t, err, errMsg)
	assert.Nil(t, v)

	v, err = f.Required(42, nil)
	assert.Error(t, err)
	assert.EqualError(t, err, "at <1>: expected string; found int")
	assert.Nil(t, v)

	v, err = f.Required()
	assert.Error(t, err)
	assert.EqualError(t, err, "wrong number of args: want 1 or 2, got 0")
	assert.Nil(t, v)

	v, err = f.Required("", 2, 3)
	assert.Error(t, err)
	assert.EqualError(t, err, "wrong number of args: want 1 or 2, got 3")
	assert.Nil(t, v)

	v, err = f.Required(0)
	require.NoError(t, err)
	assert.Equal(t, v, 0)

	v, err = f.Required("foo")
	require.NoError(t, err)
	assert.Equal(t, v, "foo")
}

func TestTernary(t *testing.T) {
	t.Parallel()

	f := TestNS()
	testdata := []struct {
		tval, fval, b interface{}
		expected      interface{}
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

	f := TestNS()
	testdata := []struct {
		arg      interface{}
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

	f := TestNS()
	truedata := []struct {
		arg  interface{}
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
		arg  interface{}
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
