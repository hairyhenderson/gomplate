package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAssert(t *testing.T) {
	_, err := Assert(false, "")
	assert.Error(t, err)
	_, err = Assert(false, "a message")
	assert.EqualError(t, err, "assertion failed: a message")

	_, err = Assert(true, "")
	assert.NoError(t, err)
}

func TestFail(t *testing.T) {
	err := Fail("")
	assert.Error(t, err)
	err = Fail("msg")
	assert.EqualError(t, err, "template generation failed: msg")
}
