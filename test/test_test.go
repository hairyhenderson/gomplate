package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssert(t *testing.T) {
	_, err := Assert(false, "")
	require.Error(t, err)
	_, err = Assert(false, "a message")
	require.EqualError(t, err, "assertion failed: a message")

	_, err = Assert(true, "")
	require.NoError(t, err)
}

func TestFail(t *testing.T) {
	err := Fail("")
	require.Error(t, err)
	err = Fail("msg")
	require.EqualError(t, err, "template generation failed: msg")
}

func TestRequired(t *testing.T) {
	v, err := Required("", nil)
	require.Error(t, err)
	assert.Nil(t, v)

	v, err = Required("", "")
	require.Error(t, err)
	assert.Nil(t, v)

	v, err = Required("foo", "")
	require.Error(t, err)
	require.EqualError(t, err, "foo")
	assert.Nil(t, v)

	v, err = Required("", 0)
	require.NoError(t, err)
	assert.Zero(t, v)

	v, err = Required("", false)
	require.NoError(t, err)
	assert.Equal(t, false, v)

	v, err = Required("", map[string]string{})
	require.NoError(t, err)
	assert.NotNil(t, v)
}
