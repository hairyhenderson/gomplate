package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssert(t *testing.T) {
	_, err := Assert(false, "")
	assert.Error(t, err)
	_, err = Assert(false, "a message")
	assert.EqualError(t, err, "assertion failed: a message")

	_, err = Assert(true, "")
	require.NoError(t, err)
}

func TestFail(t *testing.T) {
	err := Fail("")
	assert.Error(t, err)
	err = Fail("msg")
	assert.EqualError(t, err, "template generation failed: msg")
}

func TestRequired(t *testing.T) {
	v, err := Required("", nil)
	assert.Error(t, err)
	assert.Nil(t, v)

	v, err = Required("", "")
	assert.Error(t, err)
	assert.Nil(t, v)

	v, err = Required("foo", "")
	assert.Error(t, err)
	assert.EqualError(t, err, "foo")
	assert.Nil(t, v)

	v, err = Required("", 0)
	require.NoError(t, err)
	assert.Equal(t, v, 0)

	v, err = Required("", false)
	require.NoError(t, err)
	assert.Equal(t, v, false)

	v, err = Required("", map[string]string{})
	require.NoError(t, err)
	assert.NotNil(t, v)
}
