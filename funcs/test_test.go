package funcs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAssert(t *testing.T) {
	f := TestNS()
	_, err := f.Assert(false)
	assert.Error(t, err)

	_, err = f.Assert(true)
	assert.NoError(t, err)

	_, err = f.Assert("foo", true)
	assert.NoError(t, err)

	_, err = f.Assert("foo", "false")
	assert.EqualError(t, err, "assertion failed: foo")
}

func TestRequired(t *testing.T) {
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
	assert.NoError(t, err)
	assert.Equal(t, v, 0)

	v, err = f.Required("foo")
	assert.NoError(t, err)
	assert.Equal(t, v, "foo")
}
